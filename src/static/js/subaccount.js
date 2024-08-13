let subaccounts = [];

function renderSubaccounts() {
    const tableBody = document.getElementById('subaccountTableBody');
    const youare = document.getElementById('youare').innerText;
    tableBody.innerHTML = '';
    subaccounts.forEach(account => {
        const row = `
            <tr>
                <td>${account.accountname}</td>
                <td>
                    <!-- button class="action-button edit-button" onclick="editSubaccount('${account.accountname}')">修改</button>
                    <button class="action-button delete-button" onclick="deleteSubaccount('${account.accountname}')">刪除</button -->
                    <button class="action-button delete-button" onclick="switchSubaccount('${account.accountname}')">切換</button>
                </td>
            </tr>
        `;
        if (youare != account.subid){
            tableBody.innerHTML += row;
        }
    });
}
function fetchSubaccountData() {
    fetch('/subaccount/list')
        .then(response => {
            if (response.ok) {
                response.json().then(data => {
                    if (!response.ok) {
                        return;
                    } else {
                        subaccounts = []
                        if (data.data != null) {
                            data.data.forEach(subacc => {
                                subaccounts.push({
                                    accountname: subacc.accountname,
                                    refid: subacc.refid,
                                    subid: subacc.subid
                                });
                            })
                            renderSubaccounts()
                        }
                    }
                }).catch(error => {
                    console.error('Error fetching Subaccount data:', error);
                });
            }
        })
}

function addSubaccount() {
    const data = {
        'accountname': document.getElementById('newSubaccountName').value
    };
    fetch(`/subaccount`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    }).then(response => {
        if (!response.ok) {
            response.text().then(x => alert(`Failed to update Subaccount (${data.accountname}): ` + JSON.parse(x).error))
        } else {
            response.json().then(x => {
                subaccounts.push({
                    accountname: x.data.accountname,
                    refid: x.data.refid
                });
                renderSubaccounts();
                document.getElementById('newSubaccountName').value = '';
            })
        }
    }).catch(error => {
        console.error('Error:', error);
    });
}

function editSubaccount(id) {
    const account = subaccounts.find(a => a.accountname === id);
    if (account) {
        document.getElementById('editSubaccountName').value = account.accountname;
        document.getElementById('editsubaccountModal').style.display = 'block';
        document.getElementById('editsubaccountModal').dataset.refid = account.refid;
    }
}

function saveEditSubaccount() {
    const refid = document.getElementById('editsubaccountModal').dataset.refid;
    const newName = document.getElementById('editSubaccountName').value;
    const data = subaccounts.find(a => a.refid === refid);
    if (data && newName) {
        data.accountname = newName;
        fetch('/subaccount', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(response => {
            if (response.ok) {
                renderSubaccounts();
                closeEditModal();
            } else {
                alert(response.json())
            }
        })
    }
}

function deleteSubaccount(id) {
    if (confirm(`確定要刪除這個子帳號: ${id} 嗎？`)) {
        const data = subaccounts.find(a => a.accountname === id);
        fetch('/subaccount', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(response => {
            if (response.ok) {
                subaccounts = subaccounts.filter(a => a.accountname !== id);
                renderSubaccounts();
            } else {
                alert(response.json())
            }
        })

    }
}

function closeEditModal() {
    document.getElementById('editsubaccountModal').style.display = 'none';
}

function switchSubaccount(id) {
    const data = subaccounts.find(a => a.accountname === id);
    if (data) {
        fetch('/subaccount/switch', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(response => {
            if (response.ok) {
                window.location.href = '/';
            } else {
                alert(response.json())
            }
        })
    }
}

function switchback() {
    data = {
        accountname: "_MAIN_",
        refid: "_MAIN_",
    }
    fetch('/subaccount/switch', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    }).then(response => {
        if (response.ok) {
            window.location.href = '/';
        } else {
            alert(response.json())
        }
    })
}
// 初始化頁面
document.addEventListener('DOMContentLoaded', () => {
    fetchSubaccountData()
    renderSubaccounts();
});