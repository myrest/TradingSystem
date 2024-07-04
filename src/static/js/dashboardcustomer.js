let coinData = []

// 關閉模態框
function closeModal() {
    document.getElementById('cryptoModal').style.display = 'none';
}

function fetchSymbolData() {
    fetch('/customers/symbol')
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
                <td><span class="status-toggle ${item.status ? '' : 'disabled'}" onclick="toggleStatus('${item.symbol}')">${item.status ? '啟用' : '停用'}</span></td>
                <td><input type="text" class="amount-input" name="amount-${item.symbol}" value="${item.amount || 0}" ${item.SystemStatus} onchange="updateAmount('${item.symbol}')"></td>
            </tr>
        `;
        tableBody.innerHTML += row;
    });
}

function _updateCustomerSymbol(data){
    fetch(`/customers/symbol`, {
        method: 'PATCH',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    }).then(response => {
        if (!response.ok) {
            response.text().then(x => alert(`Failed to update Symbol (${Symbol}): ` + JSON.parse(x).error))
        } else {
            response.text().then(x => {
                renderCryptoTable();                    
                if (x != "\"\"") alert("Update completed, but " + x);
            })
        }
    }).catch(error => {
        console.error('Error:', error);
    });        
}

// 切換狀態
function updateAmount(Symbol) {
    const crypto = coinData.find(item => item.symbol === Symbol);
    const amount = document.getElementsByName('amount-' + Symbol)[0].value;
    if (crypto) {
        crypto.amount = amount

        const customersymbol = {
            'symbol': Symbol,
            'status': crypto.status.toString(),
            'amount': amount
        };
        _updateCustomerSymbol(customersymbol)
    }
}

// 切換狀態
function toggleStatus(Symbol) {
    const crypto = coinData.find(item => item.symbol === Symbol);
    const amount = document.getElementsByName('amount-' + Symbol)[0].value;
    if (crypto) {
        crypto.status = crypto.status ? false : true;
        crypto.amount = amount

        const customersymbol = {
            'symbol': Symbol,
            'status': crypto.status.toString(),
            'amount': amount
        };
        _updateCustomerSymbol(customersymbol)
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
