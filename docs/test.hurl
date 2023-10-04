POST https://token-exchange-api.dimo.zone/v1/tokens/exchange
Content-Type: application/json
Authorization: Bearer {{auth-token}}
{
    "nftContractAddress": "0xbA5738a18d83D41847dfFbDC6101d37C69c9B0cF",
    "tokenId": {{token-id}},
    "privileges": [4]
}
HTTP 200
[Captures]
privilege-token: jsonpath "$['token']"

GET https://trips-api.dimo.zone/v1/vehicle/{{token-id}}/trips
Authorization: Bearer {{privilege-token}}
HTTP 200
