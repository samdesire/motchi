import Navbar from "../components/Navbar"
import { NavLink } from "react-router-dom";

import styles from './Styles/home.module.css'

import pet from '../assets/cactee.svg'
import game_icon from '../assets/game_icon.svg'
import pets_icon from '../assets/pets.svg'
import shop_icon from '../assets/shop.svg'
import person_icon from '../assets/person_icon.svg'

function Home() {
    return (
        <>
            <Navbar />

            <main className={`${styles.gameContainer}`}>
            {/* <!-- Left column --> */}
                <div className={`${styles.sideColumn}`}>
                    <NavLink to='/profile'>
                        <button className={`${styles.actionBtn} ${styles.personBtn}`}>
                                <img src={`${person_icon}`} alt="" className={`${styles.personIcon} ${styles.gameIcon}`} />
                        </button>
                    </NavLink>
                    <button className={`${styles.actionBtn} ${styles.petsBtn}`}>
                        <NavLink to='/pets'>
                            <img src={`${pets_icon}`} alt="" className={`${styles.petsIcon} ${styles.gameIcon}`} />
                        </NavLink>
                    </button>
                </div>
{/* 
            <!-- Center column --> */}
                <div className={`${styles.centerColumn}`}>
                    <img src={`${pet}`} alt="" className={`${styles.petImg}`} />
                </div>

            {/* <!-- Right column --> */}
                <div className={`${styles.sideColumn}`}>
                    <button className={`${styles.actionBtn} ${styles.shopBtn}`}>
                        <NavLink to='/shop'>
                            <img src={`${shop_icon}`} alt="" className={`${styles.shopIcon} ${styles.gameIcon}`} />
                        </NavLink>
                    </button>
                    <button className={`${styles.actionBtn} ${styles.gameBtn}`}>
                        <NavLink to='/mingames'>
                            <img src={`${game_icon}`} alt="" className={`${styles.gameIcon}`} />
                        </NavLink>
                    </button>
                </div>
            </main>
        </>
    );
}

export default Home;