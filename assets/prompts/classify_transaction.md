<context_gathering>
Goal: Given information about credit card or bank transactions, you should classify which provided category the transaction belongs to.
Method:

- Look at provided categories
- Classify the type of transaction
  - Use all information provided, including the description, the merchant, the time of year, and the amount
- Look up the transaction description or merchant if you are not already instantly aware of the category the transaction fits into
- Be as efficient as possible
  Details:
- price amount is in canadian dollars
- this is for a budgeting app
- the input will be in XML to easily understand
- output the results in JSON
  Example Input:

```xml
<input>
  <categories>
    <category>
      <id>
        123
      </id>
      <name>
        Groceries
      </name>
    </category>
    <category>
      <id>
        1
      </id>
      <name>
        Mortgage
      </name>
    </category>
    <category>
      <id>
        2
      </id>
      <name>
        Utilities
      </name>
    </category>
    <category>
      <id>
        3
      </id>
      <name>
        Entertainment
      </name>
    </category>
     <category>
      <id>
        ...
      </id>
      <name>
        ...
      </name>
    </category>
    ...
  </categories>
  <transactions>
    <transaction>
      <id>
        112233
      </id>
      <description>
        FRIAR & FIRKIN TORONTO
      </description>
      <merchant>
        RESTAURANT TRANSACTION
      </merchant>
      <date>
        2025-09-16
      </date>
      <amount>
        $27.95
      </amount>
    </transaction>
    <transaction>
      <id>
        12345
      </id>
      <description>
      BMO FOOD & BEVERAGE TORONTO
      </description>
      <merchant>
        BMO FOOD BEVERAGES
      </merchant>
      <date>
        2025-09-13
      </date>
      <amount>
        $7.91
      </amount>
    </transaction>
    <transaction>
      <id>
        ...
      </id>
      <description>
      ...
      </description>
      <merchant>
        ...
      </merchant>
      <date>
        ...
      </date>
      <amount>
        ...
      </amount>
    </transaction>
  </transactions>
  ...
</input>
```

Example Output:

```json
[
  transactions: [
    {
      id: 112233,
      categoryId: 3
    },
    {
      id: 12345,
      categoryId: 3
    },
    {
      id: ...,
      categoryId: ...
    },
    ...
  ]
]
```

</context_gathering>
