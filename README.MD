## Интеграция Wordpress c Booking Addon for WooCommerce от Wordpress и системы r-keeper

## Проблемы

* pc.Parent = cfg.WOOCOMMERCE.MenuCategoryId // TODO новая папка создается в этой категории по конфигу
* ??? надо ли так делать, может если найдена то ничего не делать и по второму этапу синхронизации все присвоить?????? pc.Parent = cfg.WOOCOMMERCE.MenuCategoryId // TODO если папка найдена, то она присваивается к ккегории по умолчанию
* у продукта, который мы создали в woo должен быть уникальный идентификатор из кипера, чтобы я могу потом его найти, а не создавать новые папки, а менять его роддительскую папку и другие параметры(имя и parent только)

## Решение
// TODO поиск надо делать по уникальному идентификатору из кипера, потому что сперва происходит создание продукта на основе кипера и потом попытка присвоить идентификатор
// и если идентификатор не присвоился в кипере, то блюдо надо будет !!создавать снова!! потому что оно не найдено
// пока сделать в description добавить идентификатор и по нему искать в productCategoriesMapByID


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

### Параметры RK7
* Один заказ на визит - Включен
* Пустые визиты - Выключен

### "Костыль" Win1251 кодировка
#### Алгоритм реализации:

Скрипт, при открытии заказа на кассе rkeeper, отправляет Get запрос в микросервис с параметром name. В ответе скрипт получает значение поля name и сохраняет его в расширенном свойстве с именем name1521. МС всегда игнорирует поля с именем name1251 и в результате не получает ошибку при попытке преобразовать win1251<>utf8.

* Скрипт на форме Редактирование заказа
    * Событие - CheckViewOnShow
    * Код:
```
procedure saveProp(key, value: String);
var
  Props: TVisitOrderInfo;
begin
  Props := TVisitOrderInfo.Create();
  Props.CustomProps.Values[key+'1251'] := value;
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
            ErrorCode := HttpGet('http://192.168.0.11:8081/prop/' + inttostr(RKCheck.CurrentOrder.UVN) + '/' + prop.propname , value);      
            if ErrorCode = 0 then
               begin 
               saveProp(prop.propname, value);
               end; 
            end;   
    end;
end;

procedure CheckViewOnShowScript(Sender: TObject);
var  R : string;
  ErrorCode : integer;
begin
 updateProps();
  updateProps();
 
end;
```

#### Расширенные свойства

#### 1. Пользовательские, для элементов: Меню, Группа Меню 
* genWooID
* genWooParentCategoryID

#### 2. Свойства для заказов:
##### Включенные для отображения
* CompanyName1251 - Компания
* HallName1251 - Зал
* OrderDetails1251 - Примечание
* PersonName1251 - Имя клиента
* Phone - Телефон
* Email - Почта
##### Отключенные для отображения
* LastName1251 - Фамилия
* CompanyName - Компания
* HallName - Зал
* OrderDetails - Примечание
* PersonName - Имя клиента

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

    