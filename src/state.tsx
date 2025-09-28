export function initLocalStorage() {
    if (localStorage.getItem('money') === null) {
        localStorage.setItem('money', '10');
    } 
    if (localStorage.getItem('happiness') === null) {
        localStorage.setItem('happiness', '50');
    }
    if (localStorage.getItem('hunger') === null) {
        localStorage.setItem('hunger', '50');
    }
    if (localStorage.getItem('health') === null) {
        localStorage.setItem('health', '50');
    }
    if (localStorage.getItem('pet') === null) {
        localStorage.setItem('pet', "none");
    }
}

export function addHappiness(amount: number) {
    const happiness = localStorage.getItem('happiness');
    if (happiness !== null && parseInt(happiness) + amount <= 100) {
        localStorage.setItem('happiness', (parseInt(happiness) + 1).toString());
    } else {
        localStorage.setItem('happiness', '100');
    }
}

export function addHunger(amount: number) {
    const hunger = localStorage.getItem('hunger');
    if (hunger !== null && parseInt(hunger) + amount <= 100) {
        localStorage.setItem('hunger', (parseInt(hunger) + 1).toString());
    } else {
        localStorage.setItem('hunger', '100');
    }
}

export function addHealth(amount: number) {
    const health = localStorage.getItem('health');
    if (health !== null && parseInt(health) + amount <= 100) {
        localStorage.setItem('health', (parseInt(health) + amount).toString());
    } else {
        localStorage.setItem('health', '100');
    }
}

export function spendMoney(amount: number): boolean {
    const money = localStorage.getItem('money');

    if (money !== null && parseInt(money) >= amount) {
        localStorage.setItem('money', (parseInt(money) - amount).toString());
        return true;
    }
    return false;
}

export function earnMoney(amount: number) {
    const money = localStorage.getItem('money');
    if (money !== null) {
        localStorage.setItem('money', (parseInt(money) + amount).toString());
    } else {
        localStorage.setItem('money', amount.toString());
    }
}

export function changePet(pet: string) {
    localStorage.setItem('pet', pet);
}
