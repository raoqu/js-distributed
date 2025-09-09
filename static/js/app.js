// Wait for DOM to be fully loaded
document.addEventListener('DOMContentLoaded', function() {
    // DOM Elements
    const taskList = document.getElementById('task-list');
    const taskNameInput = document.getElementById('task-name');
    const newTaskBtn = document.getElementById('new-task-btn');
    const importBtn = document.getElementById('import-btn');
    const exportBtn = document.getElementById('export-btn');
    const saveBtn = document.getElementById('save-btn');
    const deleteBtn = document.getElementById('delete-btn');
    const runBtn = document.getElementById('run-btn');
    const browseBtn = document.getElementById('browse-btn');
    const modal = document.getElementById('modal');
    const importModal = document.getElementById('import-modal');
    const closeModal = document.querySelector('.close');
    const closeImportModal = document.querySelector('.import-close');
    const newTaskNameInput = document.getElementById('new-task-name');
    const createTaskBtn = document.getElementById('create-task-btn');
    const importForm = document.getElementById('import-form');
    
    // Global variables
    let editor;
    let currentTask = null;
    
    // Initialize Monaco Editor
    require(['vs/editor/editor.main'], function() {
        editor = monaco.editor.create(document.getElementById('editor'), {
            value: '// JavaScript task script\n\n',
            language: 'javascript',
            theme: 'vs-dark',
            automaticLayout: true,
            minimap: {
                enabled: true
            },
            scrollBeyondLastLine: true, // Allow scrolling beyond last line for better visibility
            fontSize: 14,
            lineNumbers: 'on',
            renderLineHighlight: 'all',
            formatOnType: true,
            formatOnPaste: true,
            scrollbar: {
                vertical: 'visible',
                horizontal: 'visible',
                verticalScrollbarSize: 16,
                horizontalScrollbarSize: 16,
                alwaysConsumeMouseWheel: false
            }
        });
        
        // Handle window resize to ensure editor layout is updated
        window.addEventListener('resize', function() {
            if (editor) {
                editor.layout();
            }
        });
        
        // Add keyboard shortcuts - attach to window to ensure it works globally
        window.addEventListener('keydown', function(e) {
            // Check for modifier keys (Ctrl on Windows/Linux or Command on macOS)
            const modifierKey = e.ctrlKey || e.metaKey; // metaKey is Command key on macOS
            
            // Modifier+S for Save
            if (modifierKey && (e.key === 's' || e.key === 'S' || e.keyCode === 83)) {
                // Only handle if a task is selected
                if (currentTask) {
                    e.preventDefault(); // Prevent browser's save dialog
                    e.stopPropagation(); // Stop event bubbling
                    saveBtn.click();
                    
                    return false;
                }
            }
            
            // Modifier+Shift+R for Run
            if (modifierKey && e.shiftKey && (e.key === 'r' || e.key === 'R' || e.keyCode === 82)) {
                // Only handle if a task is selected
                if (currentTask) {
                    e.preventDefault(); // Prevent browser's refresh
                    e.stopPropagation(); // Stop event bubbling
                    runBtn.click();
                    
                    return false;
                }
            }
        }, true); // Use capture phase to intercept events before they reach other handlers
        
        // Load task list after editor is initialized
        loadTaskList();
    });
    
    // API Functions
    async function importScripts(formData) {
        try {
            // Show notification that import is starting
            showNotification('正在导入脚本...', 'info');
            
            const response = await fetch('/manage/import', {
                method: 'POST',
                body: formData
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error! Status: ${response.status}`);
            }
            
            const data = await response.json();
            
            // Show success notification
            showNotification(data.message, 'success');
            
            // Reload task list to show imported scripts
            await loadTaskList();
            
            return data;
        } catch (error) {
            console.error('Error importing scripts:', error);
            showNotification(`导入失败: ${error.message}`, 'error');
            return null;
        }
    }
    
    async function exportScripts() {
        try {
            // Show notification that export is starting
            showNotification('正在导出脚本...', 'info');
            
            // Create a link element to trigger the download
            const link = document.createElement('a');
            link.href = '/manage/export';
            link.download = 'task-scripts.zip';
            
            // Append to body, click and remove
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            
            // Show success notification after a short delay
            setTimeout(() => {
                showNotification('脚本导出成功', 'success');
            }, 1000);
        } catch (error) {
            console.error('Error exporting scripts:', error);
            showNotification('导出失败', 'error');
        }
    }
    
    async function loadTaskList() {
        try {
            const response = await fetch(`/scripts`);
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            
            const data = await response.json();
            renderTaskList(data.tasks || []);
        } catch (error) {
            showNotification('加载失败', 'error');
        }
    }
    
    async function loadTaskScript(taskName) {
        try {
            const response = await fetch(`/scripts/${taskName}`);
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            
            showNotification(`加载失败: ${taskName}`, 'error');
            return null;
        }
    }
    
    async function saveTaskScript(taskName, code) {
        try {
            const response = await fetch(`/scripts/${taskName}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    name: taskName,
                    code: code
                })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || `HTTP error! Status: ${response.status}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error(`Error saving task script ${taskName}:`, error);
            showNotification(`保存失败: ${error.message}`, 'error');
            return null;
        }
    }
    
    async function deleteTaskScript(taskName) {
        try {
            const response = await fetch(`/scripts/${taskName}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            
            const data = await response.json();
            return data;
        } catch (error) {
            console.error(`Error deleting task script ${taskName}:`, error);
            showNotification(`删除失败: ${taskName}`, 'error');
            return null;
        }
    }
    
    async function executeTask(taskName) {
        try {
            // Get the endpoint from the window.appConfig (will be set by the server)
            const endpoint = window.appConfig?.scriptEndpoint || 'scripts';
            const response = await fetch(`/${endpoint}/${taskName}`);
            var data
            try {       
                data = await response.json();
            }
            catch(e) {
                throw e
            }
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            } else {
                if ( data && data.output ) {
                    // replace \n to <br/>
                    showNotification(`执行成功: ${data.output.replace(/\n/g, '<br/>')}`, 'success');
                }
                if ( data && data.console) {
                    console.log(data.console)
                }
            }
            return data;
        } catch (error) {
            console.error(`Error executing task ${taskName}:`, error);
            showNotification(`执行失败: ${taskName} ${data.error}`, 'error');
            return null;
        }
    }
    
    // UI Functions
    function renderTaskList(tasks) {
        taskList.innerHTML = '';
        
        if (tasks.length === 0) {
            const emptyItem = document.createElement('li');
            emptyItem.textContent = 'No tasks available';
            emptyItem.classList.add('empty-list');
            taskList.appendChild(emptyItem);
            return;
        }
        
        tasks.forEach(task => {
            const li = document.createElement('li');
            li.textContent = task;
            li.dataset.taskName = task;
            
            li.addEventListener('click', async () => {
                // Remove active class from all items
                document.querySelectorAll('.task-list li').forEach(item => {
                    item.classList.remove('active');
                });
                
                // Add active class to clicked item
                li.classList.add('active');
                
                // Load task script
                await selectTask(task);
            });
            
            taskList.appendChild(li);
        });
    }
    
    async function selectTask(taskName) {
        const taskData = await loadTaskScript(taskName);
        if (taskData) {
            currentTask = taskName;
            taskNameInput.value = taskName;
            editor.setValue(taskData.code || '// JavaScript task script\n\n');
            
            // Enable buttons
            taskNameInput.disabled = false;
            saveBtn.disabled = false;
            deleteBtn.disabled = false;
            runBtn.disabled = false;
            browseBtn.disabled = false;
        }
    }
    
    function clearEditor() {
        currentTask = null;
        taskNameInput.value = '';
        editor.setValue('// JavaScript task script\n\n');
        
        // Disable buttons
        taskNameInput.disabled = true;
        saveBtn.disabled = true;
        deleteBtn.disabled = true;
        runBtn.disabled = true;
        browseBtn.disabled = true;  
              
        // Remove active class from all items
        document.querySelectorAll('.task-list li').forEach(item => {
            item.classList.remove('active');
        });
    }
    
    function showNotification(message, type = 'info') {
        // Create toast container if it doesn't exist
        let toastContainer = document.querySelector('.toast-container');
        if (!toastContainer) {
            toastContainer = document.createElement('div');
            toastContainer.className = 'toast-container';
            document.body.appendChild(toastContainer);
        }
        
        // Create toast element
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        
        // Create message element
        const messageEl = document.createElement('span');
        messageEl.innerHTML = message;
        toast.appendChild(messageEl);
        
        // Create close button
        const closeBtn = document.createElement('button');
        closeBtn.className = 'toast-close';
        closeBtn.innerHTML = '&times;';
        closeBtn.addEventListener('click', () => removeToast(toast));
        toast.appendChild(closeBtn);
        
        // Add toast to container
        toastContainer.appendChild(toast);
        
        // Auto-remove after 5 seconds
        setTimeout(() => removeToast(toast), 5000);
    }
    
    function removeToast(toast) {
        if (!toast) return;
        
        // Add animation class
        toast.style.animation = 'toast-out 0.3s forwards';
        
        // Remove after animation completes
        setTimeout(() => {
            if (toast.parentNode) {
                toast.parentNode.removeChild(toast);
            }
        }, 300);
    }
    
    // Event Listeners
    importBtn.addEventListener('click', () => {
        importModal.style.display = 'block';
    });
    
    newTaskBtn.addEventListener('click', () => {
        newTaskNameInput.value = '';
        modal.style.display = 'block';
    });
    
    closeModal.addEventListener('click', () => {
        modal.style.display = 'none';
    });
    
    closeImportModal.addEventListener('click', () => {
        importModal.style.display = 'none';
    });
    
    window.addEventListener('click', (event) => {
        if (event.target === modal) {
            modal.style.display = 'none';
        } else if (event.target === importModal) {
            importModal.style.display = 'none';
        }
    });
    
    importForm.addEventListener('submit', async (event) => {
        event.preventDefault();
        const formData = new FormData(importForm);
        await importScripts(formData);
        importModal.style.display = 'none';
    });
    
    createTaskBtn.addEventListener('click', async () => {
        const taskName = newTaskNameInput.value.trim();
        
        if (!taskName) {
            showNotification('Task name cannot be empty', 'error');
            return;
        }
        
        // Create a new task script with default content
        const result = await saveTaskScript(taskName, '// JavaScript task script for ' + taskName + '\n\n');
        
        if (result) {
            modal.style.display = 'none';
            await loadTaskList();
            
            // Select the newly created task
            await selectTask(taskName);
            
            // Find and highlight the new task in the list
            const taskItems = document.querySelectorAll('.task-list li');
            taskItems.forEach(item => {
                if (item.dataset.taskName === taskName) {
                    item.classList.add('active');
                }
            });
        }
    });
    
    saveBtn.addEventListener('click', async () => {
        if (!currentTask) return;
        
        const code = editor.getValue();
        const result = await saveTaskScript(currentTask, code);
        
        if (result) {
            showNotification(`'${currentTask}' 已保存`, 'success');
        }
    });
    
    deleteBtn.addEventListener('click', async () => {
        if (!currentTask) return;
        
        // Show confirmation toast
        showDeleteConfirmation(currentTask);
    });
    
    function showDeleteConfirmation(taskName) {
        // Create toast container if it doesn't exist
        let toastContainer = document.querySelector('.toast-container');
        if (!toastContainer) {
            toastContainer = document.createElement('div');
            toastContainer.className = 'toast-container';
            document.body.appendChild(toastContainer);
        }
        
        // Create toast element
        const toast = document.createElement('div');
        toast.className = 'toast warning';
        
        // Create message element
        const messageEl = document.createElement('div');
        messageEl.style.display = 'flex';
        messageEl.style.flexDirection = 'column';
        messageEl.style.gap = '10px';
        
        const textEl = document.createElement('span');
        textEl.textContent = `Delete task script '${taskName}'?`;
        messageEl.appendChild(textEl);
        
        // Create action buttons
        const btnContainer = document.createElement('div');
        btnContainer.style.display = 'flex';
        btnContainer.style.gap = '10px';
        
        const confirmBtn = document.createElement('button');
        confirmBtn.className = 'btn btn-danger';
        confirmBtn.textContent = 'Delete';
        confirmBtn.style.padding = '4px 8px';
        confirmBtn.style.fontSize = '12px';
        confirmBtn.addEventListener('click', async () => {
            removeToast(toast);
            const result = await deleteTaskScript(taskName);
            
            if (result) {
                showNotification(`'${taskName}' 已删除`, 'success');
                await loadTaskList();
                clearEditor();
            }
        });
        
        const cancelBtn = document.createElement('button');
        cancelBtn.className = 'btn';
        cancelBtn.textContent = 'Cancel';
        cancelBtn.style.padding = '4px 8px';
        cancelBtn.style.fontSize = '12px';
        cancelBtn.style.backgroundColor = '#95a5a6';
        cancelBtn.style.color = 'white';
        cancelBtn.addEventListener('click', () => removeToast(toast));
        
        btnContainer.appendChild(confirmBtn);
        btnContainer.appendChild(cancelBtn);
        messageEl.appendChild(btnContainer);
        
        toast.appendChild(messageEl);
        
        // Add toast to container
        toastContainer.appendChild(toast);
        
        // Auto-remove after 10 seconds
        setTimeout(() => removeToast(toast), 10000);
    }
    
    runBtn.addEventListener('click', async () => {
        if (!currentTask) return;
        
        // First save the current script
        const code = editor.getValue();
        await saveTaskScript(currentTask, code);
        
        // Then execute the task
        const result = await executeTask(currentTask);
        
        if (result) {
            showNotification(`\u5f00\u59cb\u8fd0\u884c '${currentTask}'`, 'success');
        }
    });

    browseBtn.addEventListener('click', async () => {
        if (!currentTask) return;
        
        const url = `/${window.appConfig?.scriptEndpoint}/${currentTask}`;
        window.open(url, '_blank');
    });
    
    // Export button event listener
    exportBtn.addEventListener('click', () => {
        exportScripts();
    });
});
