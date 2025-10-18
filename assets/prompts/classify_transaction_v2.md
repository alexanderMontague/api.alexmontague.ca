You are classifying financial transactions into the provided categories.
If you are unsure of what category a transaction fits into by the name, research the name of the merchant/description first, then place in the "Other" category.
Return valid JSON matching this schema:

{
  "transactions": [
    { "id": string, "categoryId": string }
  ]
}

Example Input:
<input>
  <categories>
    <category>
      <id>
        a74e52df-a7f2-4920-9007-d44cee2e5d0d
      </id>
      <name>
        Dining Out
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

Example Output:

{
  transactions: [
    { "id": "e77dcf5f-d8ce-4fae-a776-266a42c9e81b", "categoryId": "a74e52df-a7f2-4920-9007-d44cee2e5d0d" }
  ]
}