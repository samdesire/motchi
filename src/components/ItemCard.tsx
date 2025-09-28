
import styles from './Styles/itemcard.module.css'

interface Props {
    itemName: string,
    itemValue: number,
    itemDescription: string,
    itemImg: string,
    itemCost: number
}

import { spendMoney } from '../state'

function ItemCard(props : Props) {
    const handleBuy = (amount : number) => {
        console.log(spendMoney(amount))
    }

    return (
        <>
            <div className={`${styles.card}`}>
                <h2>{props.itemName}</h2>
                <img src={props.itemImg} alt="potion item" className={`${styles.itemImg}`} />
                <div className={`${styles.moreInfo}`}>
                    <p>{props.itemDescription}</p>
                    <div className={`${styles.purchace}`}>
                        <p>Price: ${props.itemCost}</p>
                        <button onClick={() => handleBuy(props.itemCost)}>
                            Buy
                        </button>
                    </div>
                </div>
            </div>
        </>
    );
}

export default ItemCard