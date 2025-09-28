
import { changePet, spendMoney } from '../state';
import styles from './Styles/itemcard.module.css'

interface Props {
    petName: string,
    petDescription: string,
    petImg: string,
}

function PetCard(props : Props) {
    function onClick() {
        if(spendMoney(10)) {
            changePet(props.petImg);
            alert(`You have adopted ${props.petName} as your new Motchi!`);
        } else {
            alert("You don't have enough coins to adopt this pet!");
        }
    }

    return (
        <>
            <div className={`${styles.card}`} onClick={() => onClick()}>
                <h2>{props.petName}</h2>
                <img src={props.petImg} alt="pet image" className={`${styles.itemImg}`} />
                <div className={`${styles.moreInfo}`}>
                    <p>{props.petDescription}</p>
                </div>
            </div>
        </>
    );
}

export default PetCard