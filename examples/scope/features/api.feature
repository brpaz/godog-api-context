Feature: API Tests
  Scenario: Test GET request
    Given I set header "Content-Type" with value "application/json"
    When I send "GET" request to "/status/200"
    Then The response code should be 200
    Then I store the value of response header "X-Some-Header" as "token" in scenario scope
    Then The scope variable "token" should have value "world"

  Scenario: Test GET request
    Given I set header "Content-Type" with value "application/json"
    And I set query param "token" with value "`##token`"
    When I send "GET" request to "/status/200"
    Then The response code should be 200
