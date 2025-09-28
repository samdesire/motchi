
import styles from './Styles/itemcard.module.css'

interface Props {
    itemName: string,
    itemValue: number,
    itemDescription: string,
    itemImg: string,
    itemCost: number,
    type: string
}

import {addHappiness, addHealth, addHunger, spendMoney} from '../state.tsx'

function ItemCard(props : Props) {
    const handleBuy = (amount : number) => {
        if(spendMoney(amount)) {
            if(props.type == 'happiness') {
                addHappiness(10)    
            } else if (props.type == 'health')  {
                addHealth(10) 
            } else if (props.type == 'hunger') {
                addHunger(10)
            }
            alert("Item Purchased!")
        } else {
            alert("Not enough money!")
        }
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