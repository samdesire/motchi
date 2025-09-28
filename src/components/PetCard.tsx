
import styles from './Styles/itemcard.module.css'

interface Props {
    petName: string,
    petDescription: string,
    petImg: string,
}

function PetCard(props : Props) {
    return (
        <>
            <div className={`${styles.card}`}>
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