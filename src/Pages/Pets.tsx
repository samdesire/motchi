import './Styles/pets.css'

import PetCard from '../components/petCard';
import Navbar from '../components/Navbar';

import pinkMotchi from '../assets/pink_motchi.svg'

function Pets() {
    return (
        <>
            <Navbar />
            <div className="mainPets">
                <h1>Pets</h1>
                <PetCard petName = {"Ckerii"} petDescription={"A tiny, cherry-shaped Motchi that glows brighter the more love it receives. Playful and cuddly, Ckherri thrives on affection and spreads joy wherever it bounces."} petImg={pinkMotchi}/>
            </div>
        </>
    );
}

export default Pets;