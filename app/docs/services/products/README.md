# Products microservice documentation

## SEED(s) use cases

- Add/Get/List/Update/Delete for Product Entity
  - Filtering by *sellerId* is an important requirement for the List handler
- Question: do I extract `GetRecommendations` or similar method to the separate microservice, and add rating, paid promotion, etc. functionality to it? Probably I do (recommendation service or so?)
