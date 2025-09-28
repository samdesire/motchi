import { NavLink } from "react-router-dom";
import styles from './Styles/navbar.module.css'

import motchi_pixel_logo from '../assets/motchi_pixel_logo.svg'

import { PiBowlFood } from "react-icons/pi";
import { VscSmiley } from "react-icons/vsc";
import { IoIosHeartEmpty } from "react-icons/io";
import { BsCoin } from "react-icons/bs";

import meat from '../assets/meat.svg'
import happiness from '../assets/happiness.svg'
import heart from '../assets/health.svg'
import cash from '../assets/cash.svg'

function Navbar() {
    return (
        <nav>
            <div>
                <NavLink to='/'>
                    <img src={`${motchi_pixel_logo}`} alt="logo for motchi" className={`${styles.logo}`} />
                </NavLink>
            </div>
            <div className={`${styles.currencyDisplay}`}>
                <ul>
                    <li className={`${styles.happiness}`}>
                        <img src={happiness} alt="" className={`${styles.statusIcon}`} />
                        {/* < VscSmiley /> */}
                        <p>Happiness</p>
                    </li>
                    <li className={`${styles.hunger}`}>
                        <img src={meat} alt="" className={`${styles.statusIcon}`} />
                        {/* < PiBowlFood /> */}
                        <p>Hunger</p>
                    </li>
                    <li className={`${styles.health}`}>
                        <img src={heart} alt=""  className={`${styles.statusIcon}`}/>
                        {/* < IoIosHeartEmpty /> */}
                        <p>Health</p>
                    </li>
                    <li className={`${styles.coins}`}>
                        <img src={cash} alt="" className={`${styles.statusIcon}`} />
                        {/* < BsCoin /> */}
                        <p>$2,546</p>
                    </li>
                </ul>
            </div>
        </nav>
    )
}

export default Navbar;