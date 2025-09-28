function initLocalStorage() {
    if (localStorage.getItem('money') === null) {
        localStorage.setItem('money', '0');
    }
}

function addHappiness() {

}

function addHunger() {

}

function addHealth() {
    
}

function spendMoney(amount: number): boolean {
    const money = localStorage.getItem('money');

    if (money !== null && parseInt(money) >= amount) {
        localStorage.setItem('money', (parseInt(money) - amount).toString());
        return true;
    }
    return false;
}

export default spendMoney