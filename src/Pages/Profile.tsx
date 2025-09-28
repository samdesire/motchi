import styles from './Styles/profile.module.css'
import Navbar from '../components/Navbar';

import person_icon from '../assets/person_icon.svg'

function Profile() {
    return (
        <>
            <Navbar />
            <div className={`${styles.profileCont}`}>
                <div className={`${styles.profilePic}`}>
                    <img src={`${person_icon}`} alt="" className={`${styles.personIcon} ${styles.gameIcon}`} />
                </div>
                <div className={`${styles.profileDescriptors}`}>
                    <div className={`${styles.username} ${styles.item}`}>
                        <h2>Username</h2>
                        <p>TheAwesomeFish</p>
                    </div>
                    <div className={`${styles.partner} ${styles.item}`}>
                        <h2>Partner</h2>
                        <p>CuddleFish</p>
                    </div>
                    <div className={`${styles.pet} ${styles.item}`}>
                        <h2>Pet</h2>
                        <p>Cactee</p>
                    </div>
                    <div className={`${styles.coins} ${styles.item}`}>
                        <h2>Coins</h2>
                        <p>$2168</p>
                    </div>
                </div>
            </div>
        </>
    );
}

export default Profile