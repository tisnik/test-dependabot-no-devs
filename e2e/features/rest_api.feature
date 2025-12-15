Feature: REST API tests


  Background:
    Given The service is started locally
      And REST API service prefix is /v1

  Scenario: Check if the OpenAPI endpoint works as expected
    Given The system is in default state
     When I access endpoint "openapi.json" using HTTP GET method
     Then The status code of the response is 200
      And The body of the response contains OpenAPI
