# Order service

Request that spans across many services:

[PlantUML Editor](https://editor.plantuml.com/uml/lLGxRi904ErzYXKz5qY1a7A0S02QbRq54h6HFqKwFbAYY4YKwK2vWKrmR44mLvYzKTwk3R4T0aI9KbZMpEnxRsVcsMtFvwuVtFODRlgEosX16Mta4oLuBocKOufLpBZ70cE0ipspD-2spnethYLShw4gl5gvHg3pmgmMEgvZ1QQVWuHX1CqMlsBqJ27jv90oEM9o1E64LaXjAPKprhKHXhJ6WxV1SyXB-HJ53E6EOd24yXw95B0lJ3MUPgibp2DK6XV9ALe7gcipCAkPJEG3Pf5JwJChBVW6q3kYLyeOs3Ea4gIsSghk6a1WNhxQ7bowMAr3OUOjBkI48jg0Qrs5Ng8hJX6ezXuFlG2FOGqH9vfuYUmmtxwUICemYTmCEbrihUFq-vfEZbHTpZxtLT_5kKtaKmBB4h2nkX3EnQgpGdbRgAYr8WZPuBYu8KwFhdRLxvjMnMhJbl1yrxc3Qmi7phLbOxKDACKby_n2vS0TgPvH-4jcxVOBJQ6t664LC7sgPTVgAwlwx30JyryySSmFF_tkeSiSzUDYFUkScVRjSSDe2zl2J_83)

```plantuml
@startuml
actor Пользователь as u
participant "Orders" as a
participant "Cart" as ec
participant "Products" as e

u->a: Запрос создания \nзаказа
a->a: Создание операции \nсоздания заказа
a-->>ec: Создание события\n получения содержимого корзины
a->u: Операция создания\nзаказа
ec->a: Публикация содержимого корзины
a->e: Публикация сообщения о \nрезервировании товаров
u->a: Poll состояния \nоперации создания заказа
a->u: Ответ о неготовности \nна poll состояния
e->a: Публикация состояния резервирования товаров
destroy e
a->a: Определение состояния резервирования товаров
a->a: Обновление состояния резервирования товаров
a->ec: Публикация сообщения \nоб очистке корзины
destroy ec
u->a: Poll состояния \nоперации создания заказа
a->u: Ответ о готовности операции \nсоздания заказа с order id \nна poll состояния
destroy a

@enduml
```

