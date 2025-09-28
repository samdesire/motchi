import Navbar from "../components/Navbar"
import { NavLink } from "react-router-dom";
import { initLocalStorage } from "../state.tsx";

import styles from './Styles/home.module.css'

import pet from '../assets/cactee.svg'
import game_icon from '../assets/game_icon.svg'
import pets_icon from '../assets/pets.svg'
import shop_icon from '../assets/shop.svg'
import person_icon from '../assets/person_icon.svg'

import add_partner from '../assets/addprofile.svg'
import { useEffect, useState } from "react";

function Home() {

    const [pet, setPet] = useState('../assets/cactee.svg');
    const [money, setMoney] = useState(0);
    const [happiness, setHappiness] = useState(0);
    const [health, setHealth] = useState(0);
    const [hunger, setHunger] = useState(0);

    useEffect(() => {
        initLocalStorage();

        const storedPet = localStorage.getItem('pet');
        if (storedPet) {
            setPet(storedPet);
        }
        const storedMoney = localStorage.getItem('money');
        if (storedMoney) {
            setMoney(parseInt(storedMoney));
        }
        const storedHappiness = localStorage.getItem('happiness');
        if (storedHappiness) {
            setHappiness(parseInt(storedHappiness));
        }
        const storedHealth = localStorage.getItem('health');
        if (storedHealth) {
            setHealth(parseInt(storedHealth));
        }
        const storedHunger = localStorage.getItem('hunger');
        if (storedHunger) {
            setHunger(parseInt(storedHunger));
        }
    }, []);

    return (
        <>
            <Navbar money={money} happiness={happiness} health={health} hunger={hunger} />

            <main className={`${styles.gameContainer}`}>
            {/* <!-- Left column --> */}
                <div className={`${styles.sideColumn}`}>
                    <NavLink to='/profile'>
                        <button className={`${styles.actionBtn} ${styles.personBtn}`}>
                                <img src={`${person_icon}`} alt="" className={`${styles.personIcon} ${styles.gameIcon}`} />
                        </button>
                    </NavLink>
                    <NavLink to='/add-partner'>
                        <button className={`${styles.actionBtn} ${styles.addPartnerBtn}`}>
                                <img src={`${add_partner}`} alt="" className={`${styles.addPartnerIcon} ${styles.gameIcon}`} />
                        </button>
                    </NavLink>
                </div>
{/* 
            <!-- Center column --> */}
                <div className={`${styles.centerColumn}`}>
                    <img src={`${pet}`} alt="" className={`${styles.petImg}`} />
                    <NavLink to='/mingames'>
                        <button className={`${styles.actionBtn} ${styles.gameBtn}`}>
                                <img src={`${game_icon}`} alt="" className={`${styles.gameIcon}`} />
                        </button>
                    </NavLink>
                </div>

            {/* <!-- Right column --> */}
                <div className={`${styles.sideColumn}`}>
                    <NavLink to='/shop'>
                        <button className={`${styles.actionBtn} ${styles.shopBtn}`}>
                                <img src={`${shop_icon}`} alt="" className={`${styles.shopIcon} ${styles.gameIcon}`} />
                        </button>
                    </NavLink>
                    <NavLink to='/pets'>
                        <button className={`${styles.actionBtn} ${styles.petsBtn}`}>
                                <img src={`${pets_icon}`} alt="" className={`${styles.petsIcon} ${styles.gameIcon}`} />
                        </button>
                    </NavLink>
                </div>
            </main>
        </>
    );
}

export default Home;