import './style.css';
import {
    CreateNewInstance,
    GetInstances,
    Launch,
    DeleteInstance,
    StopInstance,
    RepairInstance,
    UpdateInstance
} from '../wailsjs/go/main/App';

// --- KILL THE WEBVIEW CONTEXT MENU ---
window.addEventListener('contextmenu', (e) => {
    if (!e.target.closest('.card')) {
        e.preventDefault();
        return false;
    }
    e.preventDefault();
}, true);

// --- DISABLE DEV SHORTCUTS ---
window.addEventListener('keydown', (e) => {
    const isCmd = e.metaKey || e.ctrlKey;
    const key = e.key.toLowerCase();

    if (
        (isCmd && key === 'r') ||
        (isCmd && e.altKey && key === 'i') ||
        key === 'f5' ||
        (isCmd && e.shiftKey && key === 'r')
    ) {
        e.preventDefault();
    }
});

// --- GLOBAL STATE ---
let selectedIconFolder = 'Icon1'; // Match exact casing
let contextTarget = { name: '', icon: '', isMain: false };

// --- INITIALIZATION ---
document.addEventListener('DOMContentLoaded', () => {
    window.updateGrid();

    document.addEventListener('click', () => {
        const menu = document.getElementById('context-menu');
        if (menu) menu.classList.add('hidden');
    });

    const menuDelete = document.getElementById('menu-delete');
    if (menuDelete) {
        menuDelete.onclick = (e) => {
            e.preventDefault();
            e.stopPropagation();
            document.getElementById('delete-instance-name').innerText = contextTarget.name;
            window.showModal('delete-modal');
            document.getElementById('context-menu').classList.add('hidden');
        };
    }

    const menuEdit = document.getElementById('menu-edit');
    if (menuEdit) {
        menuEdit.onclick = (e) => {
            e.preventDefault();
            e.stopPropagation();
            document.getElementById('edit-instance-name').value = contextTarget.name;
            window.showModal('edit-modal');
            document.getElementById('context-menu').classList.add('hidden');
        };
    }
});

// --- GRID RENDERING ---
window.updateGrid = function() {
    GetInstances().then((instances) => {
        const grid = document.getElementById('instance-grid');
        if (!grid) return;

        grid.innerHTML = `
            <div class="card create-new" onclick="window.showModal('create-modal')">
                <img src="/assets/images/instanceUIButtons/AddInstance.png" class="add-icon-img" alt="Add">
                <span class="name">Create New Instance</span>
            </div>
        `;

        const mainAppData = instances?.find(i => i.name === "Roblox") || {
            name: "Roblox",
            path: "/Applications/Roblox.app",
            status: "Operational",
            iconFolder: "Icon1"
        };

        const allInstances = instances ? [...instances] : [];
        if (!instances?.some(i => i.name === "Roblox")) {
            allInstances.unshift(mainAppData);
        }

        allInstances.forEach(inst => {
            const isMain = inst.name === "Roblox" || inst.path === "/Applications/Roblox.app";
            let displayStatus = inst.status || 'Operational';
            if (displayStatus === "Open") displayStatus = "Running";
            if (displayStatus === "UpdateRequired") displayStatus = "Broken (please repair)";

            const statusClass = (displayStatus === "Running" || displayStatus === "Operational") ? "operational" : "issue";

            // PRESERVE CASING: Ensure we use exact string from backend (Icon1, Icon2, etc.)
            const folderName = inst.iconFolder || 'Icon1';
            const iconPath = `/assets/images/instanceIcons/${folderName}.png`;

            grid.innerHTML += `
                <div class="card" 
                     data-is-main="${isMain}" 
                     oncontextmenu="window.showContextMenu(event, '${inst.name}', '${folderName}', ${isMain})">
                    <img src="${iconPath}" class="app-icon" onerror="this.src='/assets/images/appIcon/AppIcon.png'">
                    <div class="name">${inst.name}</div>
                    <div class="status ${statusClass}">${displayStatus}</div>
                    
                    <div class="card-actions">
                        <img src="/assets/images/instanceUIButtons/Play.png" class="action-btn-img" onclick="window.LaunchInstance('${inst.name}')">
                        ${!isMain ? `<img src="/assets/images/instanceUIButtons/Repair.png" class="action-btn-img" onclick="window.RepairInstance('${inst.name}')">` : ''}
                        <img src="/assets/images/instanceUIButtons/Quit.png" class="action-btn-img" onclick="window.StopInstance('${inst.name}')">
                    </div>
                </div>
            `;
        });
    }).catch(err => console.error("Could not load instances:", err));
};

// --- MODAL & OVERLAY LOGIC ---
window.showModal = function(id) {
    const overlay = document.getElementById('overlay');
    if (overlay) overlay.classList.remove('hidden');
    document.querySelectorAll('.modal').forEach(m => m.classList.add('hidden'));
    const target = document.getElementById(id);
    if (target) {
        target.classList.remove('hidden');
        // Reset selection visual to Icon1 whenever modal opens
        if (id === 'create-modal') {
            selectedIconFolder = 'Icon1';
            document.querySelectorAll('.icon-opt').forEach(opt => {
                opt.classList.remove('selected');
                if (opt.getAttribute('onclick')?.includes('Icon1')) opt.classList.add('selected');
            });
        }
    }
};

window.closeModals = function() {
    const overlay = document.getElementById('overlay');
    if (overlay) overlay.classList.add('hidden');
    document.querySelectorAll('.modal').forEach(m => m.classList.add('hidden'));
};

// --- BACKEND ACTIONS ---
window.handleCreate = function() {
    const nameInput = document.getElementById('instance-name');
    const name = nameInput.value.trim();
    if (!name) return;

    window.showNotification(`Creating ${name}...`);

    // Force a log here to see EXACTLY what is being sent to Go
    console.log("SENDING TO BACKEND:", name, selectedIconFolder);

    CreateNewInstance(name, selectedIconFolder).then(() => {
        setTimeout(() => {
            nameInput.value = "";
            window.closeModals();
            window.updateGrid();
            window.showNotification(`Created new "${name}" successfully!`);
        }, 1200);
    });
};

window.selectIcon = function(folderName, element) {
    // Force lowercase check to prevent "icon5" vs "Icon5" confusion
    const cleanName = folderName.charAt(0).toUpperCase() + folderName.slice(1);
    selectedIconFolder = cleanName;

    console.log("UI SELECTED:", selectedIconFolder);

    document.querySelectorAll('.icon-opt').forEach(opt => opt.classList.remove('selected'));
    element.classList.add('selected');
};

window.handleDelete = function() {
    const nameToDelete = contextTarget.name;
    DeleteInstance(nameToDelete).then(() => {
        window.closeModals();
        window.updateGrid();
        window.showNotification(`Deleted "${nameToDelete}" successfully!`);
    });
};

window.handleEditSave = function() {
    const newNameInput = document.getElementById('edit-instance-name');
    const newName = newNameInput.value.trim();
    if (!newName) {
        window.closeModals();
        return;
    }
    // UpdateInstance(oldName, newName, iconName)
    UpdateInstance(contextTarget.name, newName, contextTarget.icon).then(() => {
        window.closeModals();
        window.updateGrid();
    }).catch(err => console.error("Update failed:", err));
};

window.LaunchInstance = (name) => Launch(name);
window.RepairInstance = (name) => {
    RepairInstance(name).then(() => {
        window.showNotification(`Repaired "${name}" successfully!`);
    });
};
window.StopInstance = (name) => StopInstance(name);

// --- UTILS ---
window.showContextMenu = function(e, name, icon, isMain) {
    if (isMain) return;
    e.preventDefault();
    contextTarget = { name, icon, isMain }; // Sync icon for editing
    const menu = document.getElementById('context-menu');
    if (menu) {
        menu.style.top = `${e.clientY}px`;
        menu.style.left = `${e.clientX}px`;
        menu.classList.remove('hidden');
    }
};

window.selectIcon = function(folderName, element) {
    selectedIconFolder = folderName;
    document.querySelectorAll('.icon-opt').forEach(opt => opt.classList.remove('selected'));
    element.classList.add('selected');
};

window.switchTab = function(tabName, element) {
    document.querySelectorAll('.nav-item').forEach(nav => nav.classList.remove('active'));
    element.classList.add('active');
    document.querySelectorAll('.view').forEach(view => view.classList.add('hidden'));
    const targetView = document.getElementById(`view-${tabName}`);
    if (targetView) targetView.classList.remove('hidden');
    window.closeModals();
};

window.showNotification = function(message) {
    const toast = document.getElementById('notification-toast');
    if (!toast) return;

    toast.innerHTML = `<span>✅</span> ${message}`;
    toast.classList.remove('hidden');
    setTimeout(() => toast.classList.add('show'), 10);

    setTimeout(() => {
        toast.classList.remove('show');
        setTimeout(() => toast.classList.add('hidden'), 100);
    }, 2000);
};

// --- THEME SYSTEM ---
window.updateTheme = function(theme) {
    const body = document.body;

    if (theme === 'light') {
        body.classList.add('light-mode');
        body.classList.remove('dark-mode');
    } else {
        body.classList.remove('light-mode');
    }

    window.showNotification(`Switched to ${theme.charAt(0).toUpperCase() + theme.slice(1)} Mode`);
};

window.BrowserOpenURL = function(url) {
    window.go.main.App.OpenBrowser(url);
};