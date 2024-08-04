let subaccounts = [
    { id: 1, name: "子帳號1" },
    { id: 2, name: "子帳號2" },
    { id: 3, name: "子帳號3" }
];

function renderSubaccounts() {
    const tableBody = document.getElementById('subaccountTableBody');
    tableBody.innerHTML = '';
    subaccounts.forEach(account => {
        const row = `
            <tr>
                <td>${account.name}</td>
                <td>
                    <button class="action-button edit-button" onclick="editSubaccount(${account.id})">修改</button>
                    <button class="action-button delete-button" onclick="deleteSubaccount(${account.id})">刪除</button>
                </td>
            </tr>
        `;
        tableBody.innerHTML += row;
    });
}

function addSubaccount() {
    const name = document.getElementById('newSubaccountName').value;
    if (name) {
        const newId = subaccounts.length + 1;
        subaccounts.push({ id: newId, name: name });
        renderSubaccounts();
        document.getElementById('newSubaccountName').value = '';
    }
}

function editSubaccount(id) {
    const account = subaccounts.find(a => a.id === id);
    if (account) {
        document.getElementById('editSubaccountName').value = account.name;
        document.getElementById('editModal').style.display = 'block';
        document.getElementById('editModal').dataset.editId = id;
    }
}

function saveEditSubaccount() {
    const id = parseInt(document.getElementById('editModal').dataset.editId);
    const newName = document.getElementById('editSubaccountName').value;
    const account = subaccounts.find(a => a.id === id);
    if (account && newName) {
        account.name = newName;
        renderSubaccounts();
        closeEditModal();
    }
}

function deleteSubaccount(id) {
    if (confirm('確定要刪除這個子帳號嗎？')) {
        subaccounts = subaccounts.filter(a => a.id !== id);
        renderSubaccounts();
    }
}

function closeEditModal() {
    document.getElementById('editModal').style.display = 'none';
}

// 初始化頁面
document.addEventListener('DOMContentLoaded', () => {
    renderSubaccounts();
});