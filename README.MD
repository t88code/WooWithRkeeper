## Интеграция Wordpress c Booking Addon for WooCommerce от Wordpress и системы r-keeper

### Доступ

http://new.hotelslovakia.ru/wp-admin
Логин restocrm
Пароль 108restocrm108


Доступ к БД:
phpMyAdmin: https://pma.nic.ru/?hcp_servername=hotel16771.mysql
Сервер: hotel16771.mysql:3306
БД: hotel16771_hslov2
Пользователь: hotel16771_hslov
Пароль: sAKEL39_


ftp.hotel16771.nichost.ru
hotel16771_restocrm
108Resto108


http://new.hotelslovakia.ru
Пароль 1


https://docs.rkeeper.ru/


Плагин календаря:
https://wordpress.org/plugins/booking-for-woocommerce/


Пример брони с калькулятором без календаря:
http://new.hotelslovakia.ru/restoran-le-nuar/

### Сделать
1. Создать Props в RK7
   1. Примечание
   2. Сумма заказа
   3. Состав заказа
      1. Наименование позиции
      2. Цена позиции
      3. "quantity": false,
   4. ID заказа
   5. Phone
   6. Email
   7. link
   8. Дата брони
2. Заполнить Props через CreateOrder
3. Отобразить Props на кассе