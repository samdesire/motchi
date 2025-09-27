import Navbar from "../components/Navbar"

import './Styles/home.css'
import { FaUserLarge } from "react-icons/fa6";
import { MdOutlinePets } from "react-icons/md";
import { FaShop } from "react-icons/fa6";
import { FaGamepad } from "react-icons/fa";

function Home() {
    return (
        <>
            <Navbar />

            <main className="game-container">
            {/* <!-- Left column --> */}
                <div className="side-column">
                    <button className="action-btn">
                        < FaUserLarge />
                    </button>
                    <button className="action-btn">
                        <MdOutlinePets />
                    </button>
                </div>
{/* 
            <!-- Center column --> */}
                <div className="center-column">
                    Pets
                </div>

            {/* <!-- Right column --> */}
                <div className="side-column">
                    <button className="action-btn">
                        <FaShop />
                    </button>
                    <button className="action-btn">
                        <FaGamepad />
                    </button>
                </div>
            </main>
        </>
    );
}

export default Home;