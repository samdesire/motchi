
import styles from './Styles/itemcard.module.css'

interface Props {
    itemName: string,
    itemValue: number,
    itemDescription: string,
    itemImg: string,
    itemCost: number
}

function ItemCard(props : Props) {
    return (
        <>
            <div className={`${styles.card}`}>
                <h2>{props.itemName}</h2>
                <img src={props.itemImg} alt="potion item" className={`${styles.itemImg}`} />
                <div className={`${styles.moreInfo}`}>
                    <p>{props.itemDescription}</p>
                    <div className={`${styles.purchace}`}>
                        <p>Price: ${props.itemCost}</p>
                        <button>
                            Buy
                        </button>
                    </div>
                </div>
            </div>
        </>
    );
}

export default ItemCard