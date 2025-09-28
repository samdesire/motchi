
import Navbar from '../components/Navbar';
import ItemCard from '../components/ItemCard';

import potion_item from '../assets/purple_potion.svg'
import teddy_item from '../assets/teddy_bear.svg'
import apple from '../assets/apple_sprite.svg'
import style from './Styles/shop.module.css'

function Shop() {
    return(
        <>
            <Navbar />
            <div className={`${style.mainShop}`}>
                <h1>Shop</h1>
                <div className={`${style.shopItems}`}>
                    <ItemCard itemName = 'Potion' itemValue={10} itemDescription='A tiny glass vial filled with sparkly pink-red liquid. It tastes a little like cherries and makes your cheeks warm. Restores 15 HP.' itemCost={10} itemImg={potion_item} />
                    <ItemCard itemName = 'Teddy Bear' itemValue={25} itemDescription='A soft, well-loved plush bear. Its stitched smile never fades, bringing comfort to anyone who holds it. Gives 25 happiness.' itemCost={15} itemImg={teddy_item} />
                    <ItemCard itemName = 'Apple' itemValue={15} itemDescription='A shiny red apple with a sweet aroma. Crunchy, refreshing, and perfect for a quick snack. Replenishes hunger by 15.' itemCost={5} itemImg={apple} />

                </div>
            </div>

        </>
    );
}


export default Shop