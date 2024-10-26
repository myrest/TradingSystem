let coinData = []
const starageRegex = /\[(.*?)\]/;

// 展開設定區塊
function toggleCustomerSettings() {
    var header = document.querySelector('.api-keys-header');
    var content = document.querySelector('.api-keys-content');
    header.classList.toggle('active');
    if (content.style.display === "block") {
        content.style.display = "none";
    } else {
        content.style.display = "block";
    }
}

//切換實盤、模擬盤
function toggleSubscribeType(obj) {
    if (obj.innerText == "實盤") {
        obj.innerText = "模擬"
        obj.classList.add("disabled");
        document.querySelector('#amountSetting').style.visibility = 'hidden'
    } else {
        obj.classList.remove("disabled");
        obj.innerText = "實盤"
        document.querySelector('#amountSetting').style.visibility = 'visible'
    }
}

//自動跟單切換手動、自動
function toggleSubscribeStatus(obj) {
    if (obj.innerText == "停用") {
        obj.innerText = "啟用"
        obj.classList.remove("disabled");
        document.querySelector('#GroupSubscribeType').style.visibility = 'visible'
        document.querySelector('#GroupSubscribeType').style.display = 'block'
    } else {
        obj.classList.add("disabled");
        obj.innerText = "停用"
        document.querySelector('#GroupSubscribeType').style.visibility = 'hidden'
        document.querySelector('#GroupSubscribeType').style.display = 'none'
    }
}

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
        sysdisabled = ""
        if (item.SystemStatus == "Disabled") {
            sysdisabled = "SysDisabled"
        }
        if (item.simulation) {
            displayamount = 'visibilityHidden'
        } else {
            displayamount = 'displayBlock'
        }
        const matches = item.message.match(starageRegex);
        if (matches && matches.length > 1) {
            starageName = matches[1];
        } else {
            starageName = "無名"
        }
        const row = `
            <tr>
                <td>${item.symbol} <span class="info-icon" onclick="showDataModal('${item.symbol}', '${item.message.replace(/\n/g, '<br>')}')"><i class="fa-regular fa-file"></i></span></td>
                <td>${starageName}</td>
                <td><span class="status-toggle ${item.status ? '' : 'disabled'} ${sysdisabled}" onclick="updateCustomerCurrency('${item.symbol}', 'Status')">${item.status ? '啟用' : '停用'}</span></td>
                <td><span class="status-toggle ${!item.simulation ? '' : 'disabled'} ${sysdisabled}" onclick="updateCustomerCurrency('${item.symbol}', 'Simulation')">${item.simulation ? '模擬' : '實盤'}</span></td>
                <td>
                    <span class="${displayamount}" >
                        <input type="text" style="width:30%" name="amount-${item.symbol}" value="${item.amount || 0}" ${item.SystemStatus} onchange="displaySaveCurrenyBTN('${item.symbol}', 'Amount', this)">
                        X <input type="text" style="width:50px" name="leverage-${item.symbol}" value="${item.leverage || 0}" ${item.SystemStatus} onchange="displaySaveCurrenyBTN('${item.symbol}', 'Amount', this)">
                    </span>
                </td>
                <td><a href="/customers/placeorderhistory?symbol=${item.symbol}">記錄</a></td>
            </tr>
        `;
        tableBody.innerHTML += row;
    });
}

function getAvailableAmount(obj) {
    fetch(`/customers/availableamount`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        }
    }).then(response => {
        if (!response.ok) {
            response.text().then(x => alert(`Failed to get balance: ` + JSON.parse(x).error))
        } else {
            response.json().then(x => {
                obj.innerText = x.amount
                if (x.error != "") {
                    alert(x.error)
                }
            })
        }
    }).catch(error => {
        console.error('Error:', error);
    });
}

function _updateCustomerSymbol(data) {
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
var TEST
function displaySaveCurrenyBTN(Symbol, updatetype, obj) {
    TEST = obj
    parentContainer = obj.parentNode
    const hasButton = parentContainer.querySelector('button') !== null;
    if (!hasButton) {
        //<button onclick="saveCrypto()">保存</button>
        a = document.createElement("button");
        a.onclick = function () { updateCustomerCurrency(Symbol, updatetype); this.innerHTML = "儲存中" }
        a.innerHTML = "保存"
        a.classList.add("small-wide-btn");
        parentContainer.appendChild(a);
    }
}

// 更新投資金額
function updateCustomerCurrency(Symbol, updatetype) {
    const crypto = coinData.find(item => item.symbol === Symbol);
    updateleverage = "0"
    if (crypto) {
        if (crypto.SystemStatus != "Enabled") {
            alert('該幣種目前停止用。');
            return
        }
        switch (updatetype) {
            case "Amount":
                const amount = document.getElementsByName('amount-' + Symbol)[0].value;
                const leverage = document.getElementsByName('leverage-' + Symbol)[0].value < 1 ? 1 : document.getElementsByName('leverage-' + Symbol)[0].value;
                crypto.amount = amount
                crypto.leverage = leverage
                updateleverage = "1"
                break;
            case "Status":
                crypto.status = crypto.status ? false : true;
                if (crypto.status) {
                    updateleverage = "1"
                }
                break;
            default:
                crypto.simulation = crypto.simulation ? false : true;
        }
        const customersymbol = {
            'symbol': Symbol,
            'status': crypto.status.toString(),
            'amount': crypto.amount.toString(),
            'leverage': crypto.leverage.toString(),
            'simulation': crypto.simulation.toString(),
            'updateleverage': updateleverage,
        };
        _updateCustomerSymbol(customersymbol)
    }
}

// 初始化頁面
fetchSymbolData()
//renderCryptoTable();

// 窗口點擊事件，用於關閉模態框
window.onclick = function (event) {
    if (event.target == document.getElementById('cryptoModal')) {
        closeModal();
    }
}
