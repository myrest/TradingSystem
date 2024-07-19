// 登出功能
function logout() {
    document.getElementById('loader').style.display = 'block';

    fetch('/auth/google', {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        },
        body: ""
    }).then(response => {
        window.location = response.url
    }).then(data => console.log(data))
        .catch(error => console.error('Error:', error));
}

// 保存API Keys
function saveKeys() {
    const apiKey = document.getElementById('apiKey').value;
    const secretKey = document.getElementById('secretKey').value;
    const autosubscribe = document.getElementById('SubscribeStatus').innerText == "啟用";
    const subscribtype = document.getElementById('SubscribeType').innerText == "實盤";
    const amount = Number(document.getElementById('SubscribeAmount').value);
    
    const data = {
        apiKey,
        secretKey,
        autosubscribe,
        subscribtype,
        amount
    };

    fetch('/customers/update', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    }).then(response => {
        if (response.ok) {
            alert('Customer updated successfully');
        } else {
            response.json().then(error => {
                alert('Error updating customer:', error);
            });
        }
    }).catch(error => {
        alert('Error:', error);
    });
}

// 顯示資料內容模態框
function showDataModal(coin, data) {
    document.getElementById('dataModalTitle').textContent = coin;
    document.getElementById('dataModalContent').innerHTML = data;
    document.getElementById('dataModal').style.display = 'block';
}

// 關閉資料內容模態框
function closeDataModal() {
    document.getElementById('dataModal').style.display = 'none';
}
