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
4. Обновить Props через форму на кассе

## Настройки RK7

* Один заказ на визит - Включен
* Пустые визиты - Выключен
* Скрипт на форме Редактирование заказа
  * Событие - CheckViewOnShow
  * Код:
```
procedure saveProp(key, value: String);
var
  Props: TVisitOrderInfo;
begin
  Props := TVisitOrderInfo.Create();
  Props.CustomProps.Values[key] := value;       
  RKCheck.UpdateOrderProps(Props);
end;

procedure updateProps();
var value : string;
  ErrorCode : integer;
  Prop: TCustomProp;
  i: integer;
begin
  for i := 0 to RKCheck.CurrentOrder.CustomProps.Count - 1 do
    begin
    Prop := TCustomProp(RKCheck.CurrentOrder.CustomProps.Items[i]);
    if (prop.propname='HallName') or (prop.propname='CompanyName') or (prop.propname='PersonName') or (prop.propname='LastName') or (prop.propname='OrderDetails') then 
            begin       
            //gui.showmessage(Prop.propname + ' ' + Prop.Value);
            ErrorCode := HttpGet('http://192.168.0.11:8081/prop/' + inttostr(RKCheck.CurrentOrder.UVN) + '/' + prop.propname , value); 
            //gui.showmessage('error code:' + inttostr(ErrorCode) + inttostr(RKCheck.CurrentOrder.UVN) + prop.propname + value);       
            if ErrorCode = 0 then
               begin 
               saveProp(prop.propname, value);
               //gui.showmessage('обновлено:' + inttostr(RKCheck.CurrentOrder.UVN) + prop.propname + value);
               end; 
            //gui.showmessage('error: ' + inttostr(ErrorCode));
            //gui.showmessage('value' + value);
            end;   
    end;
end;


procedure CheckViewOnShowScript(Sender: TObject);
var  R : string;
  ErrorCode : integer;
begin
 updateProps();
end;
```



## Бронирование залов
### Наименовение полей

* Наименование зала
  * HallName = WebhookCreatOrder.LineItems[0].Name
  * Props - HallName
* Дата бронирования
  * DateStart = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Start.Date
  * Props - DateStart
  * time.Format()
* Время начала пользования залом 
  * TimeStart = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Start.Date
  * Props - TimeStart
  * time.Format()
* Время окончания пользования залом
  * TimeEnd = WebhookCreatOrder.LineItems[0].MetaData[0].Value.End.Date 
  * Props - TimeEnd
  * time.Format()
* Кол-во гостей
  * Persons = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Persons
  * Props - Persons
* Тип заказчика (Юр.лицо/Физ.лицо)
  * PersonType = WebhookCreatOrder.LineItems[0].MetaData[0].Value.PersonType // нет такого
  * Props - PersonType
* Имя заказчика
  * PersonName = WebhookCreatOrder.Billing.FirstName
  * Props - PersonName
* Наименование Юр.лица (если выбрано Юр.лицо)
  * CompanyName string = WebhookCreatOrder.Billing.Company
  * Props - CompanyName
* Реквизиты Юр.лица (если выбрано Юр.лицо)
  * CompanyDetails string = WebhookCreatOrder.LineItems[0].MetaData[0].Value.CompanyDetails // нет такого
  * Props - CompanyDetails
* Телефон заказчика
  * Phone string = WebhookCreatOrder.Billing.Phone 
  * Props - Phone
* e-mail заказчика
  * Email string = WebhookCreatOrder.Billing.Email 
  * Props - Email
* Комментарий
  * Comment string = WebhookCreatOrder.LineItems[0].MetaData[0].Value.Comment // нет такого
  * Props - Comment
* Дополнительные параметры к заказу
  * OrderDetails string = WebhookCreatOrder.LineItems[0].MetaData[0].Value.OrderDetails // нет такого - Notation
  * Props - OrderDetails
* Итоговая стоимость заказа
  * OrderSum string = WebhookCreatOrder.Total
  * Props - OrderSum
* Дата оформления заказа
  * DateCreated string = WebhookCreatOrder.DateCreated
  * Props - DateCreated

    