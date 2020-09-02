Feature: API Tests
  Scenario: Test GET request
    Given I set header "Content-Type" with value "application/json"
    When I send "GET" request to "/status/200"
    Then The response code should be 200
