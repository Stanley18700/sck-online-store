*** Settings ***
Library    SeleniumLibrary
Test Teardown    Close All Browsers

*** Variables ***
${URL}    http://localhost/product/list
${API_URL}    http://localhost
${BROWSER}    headlesschrome
${REMOTE_HUB_URL}
${USERNAME}    user_7
${PASSWORD}    P@ssw0rd
${MINT_POINTS}    160

*** Test Cases ***
ทดสอบ ใช้แต้มสะสม 160 แต้มเป็นส่วนลด 80.00 บาท สั่งซื้อสินค้า Balance Training Bicycle จัดส่งด้วย Kerry ชำระเงินด้วยบัตรเครดิต Visa สำเร็จ
    เตรียมแต้มสะสมให้ผู้ใช้ผ่าน API    ${USERNAME}    ${PASSWORD}    ${MINT_POINTS}
    เข้าสู่เว็บไซต์ และตรวจสอบว่า redirect มาที่    /auth/login    login-page
    เข้าสู่ระบบ    login-username-input    ${USERNAME}    login-password-input    ${PASSWORD}
    เลือกดูสินค้า    product-card-name-1    Balance Training Bicycle
    เพิ่มสินค้าลงตะกร้า
    ไปที่หน้า Checkout
    ใส่ที่อยู่จัดส่งสินค้า
    ...    ณัฐพล    ศรีสมบัติ
    ...    43/8 หมู่บ้านเปี่ยมสุข ถนนลาดพร้าว ซอย 63    กรุงเทพมหานคร
    ...    เขตวังทองหลาง    วังทองหลาง
    ...    10310    0891234567
    เลือกวิธีจัดส่งสินค้าเป็น    kerry
    ตรวจสอบแต้มสะสมที่แสดงในหน้า Checkout    160
    เลือกใช้แต้มสะสมเป็นส่วนลด และตรวจสอบส่วนลด    80.00
    เลือกช่องทางการชำระเงินแบบ VISA Credit Card    Nattapon Srisombat    5123 4500 0000 0008    01/39    100
    ตรวจสอบราคารวมที่ต้องชำระเงิน ต้องเท่ากันกับ    4,284.60
    ยืนยัน OTP
    ตรวจสอบหมายเลขพัสดุว่าต้องขึ้นต้นด้วย    KR
    ตรวจสอบแต้มคงเหลือผ่าน API ต้องเท่ากับ    0

*** Keywords ***
เตรียมแต้มสะสมให้ผู้ใช้ผ่าน API
    [Arguments]    ${username}    ${password}    ${amount}
    ${login}=    Evaluate    __import__('requests').post('${API_URL}/api/v1/login', json={'username': '${username}', 'password': '${password}'}, timeout=15).json()
    ${token}=    Set Variable    ${login}[access_token]
    ${resp}=    Evaluate    __import__('requests').post('${API_URL}/api/v1/point', json={'amount': ${amount}}, headers={'Authorization': 'Bearer ${token}'}, timeout=15).json()
    Should Be Equal As Integers    ${resp}[point]    ${amount}

ตรวจสอบแต้มคงเหลือผ่าน API ต้องเท่ากับ
    [Arguments]    ${expected-balance}
    ${login}=    Evaluate    __import__('requests').post('${API_URL}/api/v1/login', json={'username': '${USERNAME}', 'password': '${PASSWORD}'}, timeout=15).json()
    ${token}=    Set Variable    ${login}[access_token]
    ${resp}=    Evaluate    __import__('requests').get('${API_URL}/api/v1/point', headers={'Authorization': 'Bearer ${token}'}, timeout=15).json()
    Should Be Equal As Integers    ${resp}[point]    ${expected-balance}

เข้าสู่เว็บไซต์ และตรวจสอบว่า redirect มาที่
    [Arguments]    ${target-url}    ${target-element-locator}
    Open Browser    url=${URL}    browser=${BROWSER}    remote_url=${REMOTE_HUB_URL}
    Wait Until Location Is Not    location=${URL}
    Location Should Contain    ${target-url}
    Page Should Contain Element    id:${target-element-locator}

เข้าสู่ระบบ
    [Arguments]    ${username-input-locator}    ${username}    ${password-input-locator}    ${password}
    Wait Until Element Is Visible    id:${username-input-locator}
    Input Text    id:${username-input-locator}    ${username}
    Input Password    id:${password-input-locator}    ${password}
    Click Button    id:login-btn
    Wait Until Location Is    ${URL}
    Wait Until Element Is Visible    product-list

เลือกดูสินค้า
    [Arguments]    ${card-name-locator}    ${expected-product-name}
    Page Should Contain Element    id:${card-name-locator}
    Element Should Contain    id:${card-name-locator}    ${expected-product-name}
    Click Element    id:${card-name-locator}

เพิ่มสินค้าลงตะกร้า
    Wait Until Element Is Visible    id:product-detail-add-to-cart-btn
    Click Button    id:product-detail-add-to-cart-btn
    Wait Until Element Contains    id:header-menu-cart-badge    text=1

ไปที่หน้า Checkout
    Click Button    id:header-menu-cart-btn
    Wait Until Element Is Visible    id:shopping-cart-checkout-btn
    Click Element    id:shopping-cart-checkout-btn

ใส่ที่อยู่จัดส่งสินค้า
    [Arguments]    ${firstname}    ${lastname}
    ...    ${address}    ${province}    ${district}
    ...    ${subdistrict}    ${zipcode}    ${phone-number}
    Input Text    id:shipping-form-first-name-input    ${firstname}
    Input Text    id:shipping-form-last-name-input    ${lastname}
    Input Text    id:shipping-form-address-input    ${address}
    Select From List By Label    id:shipping-form-province-select    ${province}
    Select From List By Label    id:shipping-form-district-select    ${district}
    Select From List By Label    id:shipping-form-sub-district-select    ${subdistrict}
    Element Attribute Value Should Be    id:shipping-form-zipcode-input    value    ${zipcode}
    Input Text    id:shipping-form-mobile-input    ${phone-number}

เลือกวิธีจัดส่งสินค้าเป็น
    [Arguments]    ${method}
    &{DELIVERY_METHOD}    Create Dictionary
    ...    kerry=id:shipping-method-1-card
    ...    thai_post=id:shipping-method-2-card
    ...    lineman=id:shipping-method-3-card
    Click Element    ${DELIVERY_METHOD}[${method}]

ตรวจสอบแต้มสะสมที่แสดงในหน้า Checkout
    [Arguments]    ${expected-points}
    Wait Until Element Is Visible    id:discount-use-point-total
    Element Text Should Be    id:discount-use-point-total    ${expected-points} Points

เลือกใช้แต้มสะสมเป็นส่วนลด และตรวจสอบส่วนลด
    [Arguments]    ${expected-discount}
    Click Element    id:discount-use-point-input
    Wait Until Element Is Visible    id:order-summary-point-discount-price
    Wait Until Element Contains    id:order-summary-point-discount-price    ${expected-discount}
    Element Text Should Be    id:order-summary-point-discount-price    -฿${expected-discount}

เลือกช่องทางการชำระเงินแบบ VISA Credit Card
    [Arguments]    ${credit-card-name}    ${credit-card-number}    ${credit-card-expired-date}    ${credit-card-cvv}
    Click Element    id:payment-credit-input
    Input Text    id:payment-credit-form-fullname-input    ${credit-card-name}
    Input Text    id:payment-credit-form-card-number-input    ${credit-card-number}
    Input Text    id:payment-credit-form-expiry-input    ${credit-card-expired-date}
    Input Text    id:payment-credit-form-cvv-input    ${credit-card-cvv}

ตรวจสอบราคารวมที่ต้องชำระเงิน ต้องเท่ากันกับ
    [Arguments]    ${total-price}
    Element Should Be Visible    id:order-summary-total-payment-price
    Element Text Should Be    id:order-summary-total-payment-price    ฿${total-price}

ยืนยัน OTP
    Click Button    id:payment-now-btn
    Wait Until Element Is Visible    id:otp-input
    Click Button    Request OTP
    Input Text    id:otp-input    124532
    Click Button    OK

ตรวจสอบหมายเลขพัสดุว่าต้องขึ้นต้นด้วย
    [Arguments]    ${shipping-prefix}
    Wait Until Element Is Visible    id:order-success-tracking-id
    Element Should Contain    id:order-success-tracking-id    ${shipping-prefix}-
    ${tracking-id}=    Get Text    id:order-success-tracking-id
    Should Match Regexp    ${tracking-id}    ^${shipping-prefix}-\\d{7,9}$
