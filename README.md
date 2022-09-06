# To run the program
Run go run app.go

# To test the program
Open Postman and run http://localhost:6000/process with json input: 
{
  "side": "buy",
  "quantity": 1, 
  "price": 12.30
}
or {
  "side": "sell",
  "quantity": 1, 
  "price": 12.30
}

Explain

Asks / Bids Type

        Have below properties

        //ID: Unique id create by your API
        //Quantity: Unit of security to buy/sell.
        //Price: Price at which to execute the order eg: BUY 1 APPLE stock when the price reaches $150.
        //Side: Refers to Side Datatype i.e buy/sell
        //Timestamp: Unix epoch


Side Type

        Have two properties Sell -> iota i.e(0) and Buy will have a value of 1

Trade Type

        Trade data type have below properties

        //MakerOrderId: Order Id making this order. All limit orders are maker orders.
        //TakerOrderId: Order Id taking this order.
        //Quantity: Number of units that have been settled.
        //Price: Settlement price for a trade.
        //Timestamp: Unix epoch.
        
The system includes two steps:

Step 1: Place an order
Input 

{
  "side": "sell",
  "quantity": 1, 
  "price": 12.30
}

Step 2: Engine Processes an order

For Buy Order

Process a buy order only when the limit price is less than equal to the current sell price of a security. 
eg: Current APPLE sell price is 153 and limit price is 150, it will only execute when sell price gets below or equal to 150.

We check the price of every sell order if it is greater than we break else continue to match the order quantities and create trade 
whenever there is a settlement.


For Sell Order

We process only when the limit price is greater than the current sell price.

We check the price of every buy order if it is less than the limit price we break 
else continue to match the order quantities and create trade whenever there is a settlement.





