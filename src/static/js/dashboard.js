let coinData = []

// 顯示新增貨幣模態框
function showAddModal() {
    document.getElementById('modalTitle').textContent = '新增貨幣';
    document.getElementById('modalCoin').value = '';
    document.getElementById('modalData').value = '';
    document.getElementById('cryptoModal').style.display = 'block';
}

// 關閉模態框
function closeModal() {
    document.getElementById('cryptoModal').style.display = 'none';
}

function fetchSymbolData() {
    fetch('/restadmin/symbol')
        .then(response => {
            if (response.ok) {
                response.json().then(data => {
                    if (!response.ok) {
                        return;
                    } else {
                        coinData = data
                        renderCryptoTable()
                    }
                }).catch(error => {
                    console.error('Error fetching Symbol data:', error);
                });
            }
        })
}

// 渲染加密貨幣表格
function renderCryptoTable() {
    const tableBody = document.querySelector('#cryptoTable tbody');
    tableBody.innerHTML = '';
    coinData.forEach(item => {
        const row = `
            <tr>
                <td>${item.symbol} <span class="info-icon" onclick="showDataModal('${item.symbol}', '${item.message.replace(/\n/g, '<br>')}')"><i class="fa-regular fa-file"></i></span></td>
                <td>${item.cert}</td>
                <td><span class="status-toggle ${item.status ? '' : 'disabled'}" onclick="toggleStatus('${item.symbol}')">${item.status ? '啟用' : '停用'}</span></td>
                <td>
                    <button onclick="editCrypto('${item.symbol}')">編輯</button>
                </td>
                <td>
                    <a href="/restadmin/subscriber?symbol=${item.symbol}">訂閱者</a>
                </td>
            </tr>
        `;
        tableBody.innerHTML += row;
    });
}

// 切換狀態
function toggleStatus(symbol) {
    const crypto = coinData.find(item => item.symbol === symbol);
    if (crypto) {
        crypto.status = crypto.status ? false : true;

        const data = {
            'symbol': crypto.symbol,
            'status': crypto.status.toString()
        }
        fetch(`/restadmin/symbolStatus`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(response => {
            if (!response.ok) {
                console.error(`Failed to update status for item ${item} to ${status}`);
            }else{
                renderCryptoTable();
            }
        });
    }
}

// 保存加密貨幣數據
function saveCrypto() {
    const symbol = document.getElementById('modalCoin').value.toUpperCase();
    const message = document.getElementById('modalData').value;

    if (document.getElementById('modalTitle').textContent === '新增貨幣') {
        const data = {
            'symbol': symbol,
            'message': message
        }
        fetch('/restadmin/symbol', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(response => {
            if (response.ok) {
                response.json().then(data => {
                    coinData.push(
                        {
                            symbol: data.data.symbol,
                            status: data.data.status,
                            message: data.data.message,
                            cert : data.data.cert,
                            positionsize: '',
                        }
                    );
                })
            }
        })
    } else {
        const crypto = coinData.find(item => item.symbol === symbol);
        crypto.message = message;
        const data = {
            'symbol': symbol,
            'message': message,
            'status': crypto.status
        }
        fetch(`/restadmin/symbolMessage`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        }).then(response => {
            if (!response.ok) {
                console.error(`Failed to update status for item ${item} to ${status}`);
            }
        });
    }
    renderCryptoTable();
    closeModal();
}

// 顯示編輯貨幣模態框
function editCrypto(id) {
    const crypto = coinData.find(item => item.symbol === id);
    if (crypto) {
        document.getElementById('modalTitle').textContent = '編輯貨幣';
        document.getElementById('modalCoin').value = crypto.symbol;
        document.getElementById('modalData').value = crypto.message;
        document.getElementById('cryptoModal').style.display = 'block';
    }
}

// 初始化頁面
fetchSymbolData()
renderCryptoTable();

// 窗口點擊事件，用於關閉模態框
window.onclick = function (event) {
    if (event.target == document.getElementById('cryptoModal')) {
        closeModal();
    }
}
