import { NavLink } from "react-router-dom";
import styles from './Styles/navbar.module.css'

import motchi_pixel_logo from '../assets/motchi_pixel_logo.svg'

import { PiBowlFood } from "react-icons/pi";
import { VscSmiley } from "react-icons/vsc";
import { IoIosHeartEmpty } from "react-icons/io";
import { BsCoin } from "react-icons/bs";

interface Props {
    happiness?: number,
    hunger?: number,
    health?: number,
    money?: number,
}

function Navbar(props: Props) {
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
                        < VscSmiley />
                        <p>Happiness: {props.happiness}</p>
                    </li>
                    <li className={`${styles.hunger}`}>
                        < PiBowlFood />
                        <p>Hunger: {props.hunger}</p>
                    </li>
                    <li className={`${styles.health}`}>
                        < IoIosHeartEmpty />
                        <p>Health: {props.health}</p>
                    </li>
                    <li className={`${styles.coins}`}>
                        < BsCoin />
                        <p>Coins: {props.money}</p>
                    </li>
                </ul>
            </div>
        </nav>
    )
}

export default Navbar;