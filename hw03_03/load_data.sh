#!/bin/bash

curl -X POST http://localhost:8080/api/v1/reviews \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "reviews": [{
        "contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
        "userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
        "title": "I dunno",
        "genres": ["drama"],
        "review": "Meh!",
        "score": 60
      }]
    }
  }'

curl -X POST http://localhost:8080/api/v1/reviews \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "reviews": [{
        "contentId": "aaaaaaaa-1111-2222-3333-444444444444",
        "userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9a",
        "title": "Sonatine",
        "genres": ["drama", "crime"],
        "review": "Actualy one of the best films, would highly recomend",
        "score": 100
      }]
    }
  }'

# Create review for user B, content 1 only
curl -X POST http://localhost:8080/api/v1/reviews \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "reviews": [{
        "contentId": "937b33bf-066a-44f7-9a9b-d65071d27270",
        "userId": "2f99df7d-751c-40c9-aeea-8be8cd7bfa9b",
        "title": "I dunno",
        "genres": ["drama"],
        "review": "Meh!",
        "score": 50
      }]
    }
  }'