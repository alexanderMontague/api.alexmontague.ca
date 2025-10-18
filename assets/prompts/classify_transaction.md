<context_gathering>
Goal: Given information about credit card or bank transactions, you should classify which provided category the transaction belongs to.
Method:

- Look at provided categories
- Classify the type of transaction
  - Use all information provided, including the description, the merchant, the time of year, and the amount
- Look up the transaction description or merchant if you are not already instantly aware of the category the transaction fits into
- Any transaction that does not fit into a category well should be put in the "Other" category
- Be as efficient as possible
- Price amount is in canadian dollars
- This is for a budgeting app
- The input will be in XML to easily understand
- Output the results in JSON
- The output should _not_ include _any_ markdown elements and _only_ be valid, parsable JSON in the specified format

Example Input:

```xml
<input>
  <categories>
    <category>
      <id>
        a74e52df-a7f2-4920-9007-d44cee2e5d0d
      </id>
      <name>
        Groceries
      </name>
    </category>
    <category>
      <id>
        3361118f-1e16-4f77-b2fc-2c319efb9571
      </id>
      <name>
        Mortgage
      </name>
    </category>
    <category>
      <id>
        a2b41bf4-c1f3-4646-8df6-5602e9f8c8f9
      </id>
      <name>
        Utilities
      </name>
    </category>
    <category>
      <id>
        98d8d816-2b72-47ed-9dc1-9ed65d024d81
      </id>
      <name>
        Entertainment
      </name>
    </category>
    <category>
      <id>
        4201fb1f-78bc-4005-9eab-5874ad5a7d8d
      </id>
      <name>
        Other
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
        e77dcf5f-d8ce-4fae-a776-266a42c9e81b
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
        aa009c78-c795-4410-ac97-2882234f82c4
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
        042aa797-c714-48a5-a05d-d1ca45ef2094
      </id>
      <description>
      METRO 720 TORONTO
      </description>
      <merchant>
        METRO
      </merchant>
      <date>
        2025-09-15
      </date>
      <amount>
        $25.24
      </amount>
    </transaction>
    <transaction>
      <id>
        dc31a73f-a6d8-49ec-b3cd-a02fff36462b
      </id>
      <description>
      STAPLES STORE #53 CAMBRIDGE
      </description>
      <merchant>
        STAPLES
      </merchant>
      <date>
        2025-09-21
      </date>
      <amount>
        $1.24
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
    ...
  </transactions>
</input>
```

Example Output:

```json
[
  "transactions": [
    {
      "id": "e77dcf5f-d8ce-4fae-a776-266a42c9e81b",
      "categoryId": "98d8d816-2b72-47ed-9dc1-9ed65d024d81"
    },
    {
      "id": "aa009c78-c795-4410-ac97-2882234f82c4",
      "categoryId": "98d8d816-2b72-47ed-9dc1-9ed65d024d81"
    },
    {
      "id": "042aa797-c714-48a5-a05d-d1ca45ef2094",
      "categoryId": "a74e52df-a7f2-4920-9007-d44cee2e5d0d"
    },
    {
      "id": "dc31a73f-a6d8-49ec-b3cd-a02fff36462b",
      "categoryId": "4201fb1f-78bc-4005-9eab-5874ad5a7d8d"
    },
    {
      "id": ...,
      "categoryId": ...
    },
    ...
  ]
]
```

</context_gathering>
