package script

/**
 * 使用说明:
 * 1. NewScriptPool() 创建脚本池
 * 2. Inject() 注入宿主方法
 * 3. SetScript() 设置/更新脚本
 * 4. RunScript() / RunScriptAsync() - 同步/异步执行脚本
 */

import (
	"context"
	"errors"
	"fmt"
	"main/util"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/puzpuzpuz/xsync/v4"
)

// HostFunc：宿主方法签名，允许返回错误；错误会被抛为 JS 异常（goja.NewGoError）
type HostFunc func(rt *goja.Runtime, fc goja.FunctionCall) (goja.Value, error)

type programEntry struct {
	code    string
	program *goja.Program
	updated time.Time
}

// No longer needed as we're using map[string]interface{} directly

// Result：RunScript 的结果包装
type Result struct {
	Success  bool
	Value    any
	Err      error
	Duration time.Duration
}

// ScriptPool：脚本编译缓存 + 方法注入注册表（基于 xsync.Map）
type ScriptPool struct {
	scripts *xsync.Map[string, *programEntry]
	injects *xsync.Map[string, HostFunc]
	Store   ScriptStore
	Cache   *ScriptCache
}

// NewScriptPool 创建一个脚本池（xsync 容器替代锁）
func NewScriptPool(groupName string, redisClient *util.RedisClient) *ScriptPool {
	if redisClient == nil {
		panic("redisClient is nil")
	}

	store := NewScriptRedisStore(groupName, redisClient)
	pool := &ScriptPool{
		scripts: xsync.NewMap[string, *programEntry](),
		injects: xsync.NewMap[string, HostFunc](),
		Store:   store,
		Cache:   NewScriptCache(store),
	}
	pool.Cache.Initialize()
	return pool
}

// Inject 注入可被脚本调用的方法（线程安全，可重复调用覆盖旧实现）
func (p *ScriptPool) Inject(name string, fn HostFunc) error {
	if name == "" || fn == nil {
		return errors.New("inject: name and fn must be non-nil")
	}
	p.injects.Store(name, fn)
	return nil
}

// SetScript 设置/更新脚本源码，并编译为 Program 进行缓存（热更新）
func (p *ScriptPool) SetScript(name, code string) error {
	if name == "" {
		return errors.New("SetScript: empty name")
	}
	prog, err := goja.Compile(name, code, true)
	if err != nil {
		return fmt.Errorf("compile %q failed: %w", name, err)
	}
	p.scripts.Compute(name, func(cur *programEntry, loaded bool) (*programEntry, xsync.ComputeOp) {
		if !loaded || cur == nil {
			return &programEntry{
				code:    code,
				program: prog,
				updated: time.Now(),
			}, xsync.UpdateOp
		}
		cur.code = code
		cur.program = prog
		cur.updated = time.Now()
		return cur, xsync.UpdateOp
	})
	return nil
}

// RunScript 执行脚本，返回结果或错误
func (p *ScriptPool) RunScript(name string, opts map[string]interface{}) (*Result, error) {
	return p.runScriptWithContext(context.Background(), name, opts)
}

// RunScriptWithContext 同步执行 + 支持 ctx 取消（通过 goja.Interrupt 及时中断）
func (p *ScriptPool) RunScriptWithContext(ctx context.Context, name string, opts map[string]interface{}) (*Result, error) {
	return p.runScriptWithContext(ctx, name, opts)
}

// RunScriptAsync 异步执行，返回 Future（可 Wait/Result/Err；支持 ctx 取消）
func (p *ScriptPool) RunScriptAsync(ctx context.Context, name string, opts map[string]interface{}) *Future {
	ch := make(chan *Result, 1)
	f := &Future{ch: ch, ctx: ctx}
	go func() {
		res, err := p.runScriptWithContext(ctx, name, opts)
		if err != nil {
			// 仅在内部错误（如脚本不存在）时构造 Result
			ch <- &Result{Success: false, Err: err}
			return
		}
		ch <- res
	}()
	return f
}

// 内部统一执行逻辑
func (p *ScriptPool) runScriptWithContext(ctx context.Context, name string, opts map[string]interface{}) (*Result, error) {
	// 使用线程安全的方式加载脚本
	entry, loaded := p.scripts.Load(name)
	if !loaded || entry == nil || entry.program == nil {
		return nil, fmt.Errorf("script %q not found or not compiled", name)
	}
	prog := entry.program

	start := time.Now()
	// 每次创建新的运行时实例，确保并发安全
	rt := goja.New()

	// 预先分配对象映射空间，优化内存分配
	objMap := make(map[string]map[string]interface{}, 8) // 预分配合理的初始容量

	// 注入宿主方法（error -> JS 异常）
	p.injects.Range(func(k string, fn HostFunc) bool {
		// 封装函数，处理错误转异常
		wrapped := func(fc goja.FunctionCall) goja.Value {
			val, err := fn(rt, fc)
			if err != nil {
				panic(rt.NewGoError(err))
			}
			return val
		}

		// 检查是否有点号分隔的方法名
		parts := strings.Split(k, ".")
		if len(parts) > 1 && len(parts) <= 3 { // 支持最多 3 层嵌套
			// 处理嵌套对象，如 console.log
			objName := parts[0]
			methodName := parts[1]

			// 确保对象存在
			obj, exists := objMap[objName]
			if !exists {
				// 创建新对象并设置到运行时
				obj = make(map[string]interface{}, 8) // 预分配合理的初始容量
				objMap[objName] = obj
				rt.Set(objName, obj)
			}

			// 设置方法到对象
			obj[methodName] = wrapped
		} else {
			// 直接设置全局函数
			rt.Set(k, wrapped)
		}
		return true
	})

	injectsExtra(rt, opts)

	res := &Result{}

	// ctx 取消时中断 JS
	stop := installInterrupt(ctx, rt)
	defer stop()

	defer func() {
		res.Duration = time.Since(start)
		if r := recover(); r != nil {
			if ex, ok := r.(*goja.Exception); ok {
				res.Err = ex
			} else {
				res.Err = fmt.Errorf("panic: %v", r)
			}
			res.Success = false
		}
	}()

	v, err := rt.RunProgram(prog)
	if err != nil {
		res.Err = err
		res.Success = false
		return res, err
	}
	res.Value = v.Export()
	res.Success = true
	return res, nil
}

// 注入传入的变量和函数
func injectsExtra(rt *goja.Runtime, opts map[string]interface{}) {
	for k, v := range opts {
		// 处理嵌套对象路径，如 "data.user.name"
		parts := strings.Split(k, ".")
		if len(parts) > 1 {
			// 处理嵌套对象
			currentObj := make(map[string]interface{})
			rt.Set(parts[0], currentObj)

			// 递归创建嵌套对象
			for i := 1; i < len(parts)-1; i++ {
				nextObj := make(map[string]interface{})
				currentObj[parts[i]] = nextObj
				currentObj = nextObj
			}

			// 设置最终值
			currentObj[parts[len(parts)-1]] = v
		} else {
			// 处理函数类型
			switch fn := v.(type) {
			case HostFunc:
				// 封装函数，处理错误转异常
				wrapped := func(fc goja.FunctionCall) goja.Value {
					val, err := fn(rt, fc)
					if err != nil {
						panic(rt.NewGoError(err))
					}
					return val
				}
				rt.Set(k, wrapped)
			case func(rt *goja.Runtime, fc goja.FunctionCall) (goja.Value, error):
				// 处理原生 HostFunc 类型
				wrapped := func(fc goja.FunctionCall) goja.Value {
					val, err := fn(rt, fc)
					if err != nil {
						panic(rt.NewGoError(err))
					}
					return val
				}
				rt.Set(k, wrapped)
			case func(call goja.FunctionCall) goja.Value:
				// 处理简单的 goja 函数
				rt.Set(k, fn)
			default:
				// 其他类型直接注入
				rt.Set(k, v)
			}
		}
	}
}

// 安装 ctx -> goja.Interrupt
func installInterrupt(ctx context.Context, rt *goja.Runtime) (stop func()) {
	if ctx == nil || ctx == context.Background() {
		return func() {}
	}

	// 创建一个可取消的上下文
	interruptCtx, cancel := context.WithCancel(context.Background())

	go func() {
		select {
		case <-ctx.Done():
			// Call Interrupt directly as it's a function in newer goja versions
			rt.Interrupt(func() { panic(ctx.Err()) })
			return
		case <-interruptCtx.Done():
			// 已经被取消，直接返回
			return
		}
	}()

	// 返回取消函数
	return cancel
}
