import Navbar from "../components/Navbar"

import './Styles/home.css'


import pet from '../assets/cactee.svg'
import game_icon from '../assets/game_icon.svg'
import pets_icon from '../assets/pets.svg'
import shop_icon from '../assets/shop.svg'
import person_icon from '../assets/person_icon.svg'

function Home() {
    return (
        <>
            <Navbar />
            <main className="game-container">
            {/* <!-- Left column --> */}
                <div className="side-column">
                    <button className="action-btn person-btn">
                        <img src={`${person_icon}`} alt="" className="person-icon game-icon" />
                    </button>
                    <button className="action-btn pets-btn">
                        <img src={`${pets_icon}`} alt="" className="pets-icon game-icon" />

                    </button>
                </div>
{/* 
            <!-- Center column --> */}
                <div className="center-column">
                    <img src={`${pet}`} alt="" className="pet-img" />
                </div>

            {/* <!-- Right column --> */}
                <div className="side-column">
                    <button className="action-btn shop-btn">
                        <img src={`${shop_icon}`} alt="" className="shop-icon game-icon" />
                    </button>
                    <button className="action-btn game-btn">
                        <img src={`${game_icon}`} alt="" className="game-icon game-icon" />
                    </button>
                </div>
            </main>
        </>
    );
}

export default Home;