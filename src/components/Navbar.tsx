import { NavLink } from "react-router-dom";
import "./Styles/navbar.css"

import logo from '../../public/motchi_logo.png'

import { PiBowlFood } from "react-icons/pi";
import { VscSmiley } from "react-icons/vsc";
import { IoIosHeartEmpty } from "react-icons/io";
import { BsCoin } from "react-icons/bs";

function Navbar() {
    return (
        <nav>
            <div>
                <NavLink to='/'>
                    <img src={`${logo}`} alt="logo for motchi" className="logo" />
                </NavLink>
            </div>
            <div className="currency-display">
                <ul>
                    <li className="happiness">
                        < VscSmiley />
                        <p>Happiness</p>
                    </li>
                    <li className="hunger">
                        < PiBowlFood />
                        <p>Hunger</p>
                    </li>
                    <li className="health">
                        < IoIosHeartEmpty />
                        <p>Health</p>
                    </li>
                    <li className="coins">
                        < BsCoin />
                        <p>$2,546</p>
                    </li>
                </ul>
            </div>
        </nav>
    )
}

export default Navbar;