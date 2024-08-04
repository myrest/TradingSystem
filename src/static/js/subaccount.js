let subaccounts = [];

function fetchSubaccountData() {
    fetch('/subaccount/list')
        .then(response => {
            if (response.ok) {
                response.json().then(data => {
                    if (!response.ok) {
                        return;
                    } else {
                        subaccounts = data
                        renderCryptoTable()
                    }
                }).catch(error => {
                    console.error('Error fetching Subaccount data:', error);
                });
            }
        })
}


function renderSubaccounts() {
    const tableBody = document.getElementById('subaccountTableBody');
    tableBody.innerHTML = '';
    subaccounts.forEach(account => {
        const row = `
            <tr>
                <td>${account.name}</td>
                <td>
                    <button class="action-button edit-button" onclick="editSubaccount(${account.name})">修改</button>
                    <button class="action-button delete-button" onclick="deleteSubaccount(${account.name})">刪除</button>
                </td>
            </tr>
        `;
        tableBody.innerHTML += row;
    });
}

function addSubaccount() {
    const data = {
        'accountname': document.getElementById('newSubaccountName').value
    };    
    fetch(`/subaccount/update`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    }).then(response => {
        if (!response.ok) {
            response.text().then(x => alert(`Failed to update Subaccount (${data.accountname}): ` + JSON.parse(x).error))
        } else {
            response.text().then(x => {
                subaccounts.push({ accountname: data.accountname });
                renderSubaccounts();
                document.getElementById('newSubaccountName').value = '';
            })
        }
    }).catch(error => {
        console.error('Error:', error);
    });
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