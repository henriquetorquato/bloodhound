%%{init: {"flowchart": {"htmlLabels": false}} }%%
flowchart TD

classDef scoring fill:#75c4e6

%% Resource name level

flow_start((Program 
            Start))
  --> resource_level_start

flow_end((Program 
          End))

subgraph resource_level [Resource Name Level Evaluation]
  resource_level_is_authentication{Is Auth?}
  resource_level_is_query{Is Query?}
  resource_level_is_upload{Is Upload?}

  resource_level_start((Start))
    --> resource_level_is_authentication

  resource_level_end((End))

  %% Is Authentication?

  resource_level_is_authentication
    -- Yes --> resource_level_is_authentication_true(Add to score)
    --> resource_level_is_query

  resource_level_is_authentication
    -- No --> resource_level_is_query

  %% Is Query?

  resource_level_is_query
    -- Yes --> resource_level_is_query_true(Add to score)
    --> resource_level_is_upload

  resource_level_is_query
    -- No --> resource_level_is_upload

  %% Is Upload

  resource_level_is_upload
    -- Yes --> resource_level_is_upload_true(Add to score)
    --> resource_level_end

  resource_level_is_upload
    -- No --> resource_level_end
end

resource_level_end
  --> request_resource("`Execute GET request on resource`")
  --> request_resource_result{"`Response 
                              Successful?`"}

request_resource_result
  -- Yes --> content_level_start

request_resource_result_unsupported{Is status 
                                    code 405?}

request_resource_result
  -- No --> request_resource_result_unsupported

request_resource_result_unsupported
  -- No --> flow_end

request_resource_result_unsupported
  -- Yes --> request_resource_result_unsupported_retry(Make request 
                                                      with POST, PUT, ...)
  --> request_resource_result_unsupported_retry_result{Response 
                                                        Successful?}
  -- No --> request_resource_result_unsupported

request_resource_result_unsupported_retry_result
  -- Yes --> request_resource_result_unsupported_retry_result_true(Add to score)
  --> content_level_start

subgraph content_level [Page Content Level Evaluation]
  direction LR

  content_level_start((Start))
    --> content_level_content_type_data{Is JSON or XML?}
  
  content_level_content_type_html{Is HTML?}
  content_level_content_type_js{Is JavaScript?}
  
  content_level_send_to_js[[Send to JavaScript Evaluation]]
  content_level_send_to_html[[Send to HTML Evaluation]]

  content_level_end((End))

  %% Check for data

  content_level_content_type_data
    -- Yes --> content_level_content_type_data_true(Add to score)
    --> content_level_end

  content_level_content_type_data
    -- No --> content_level_content_type_html

  %% Check for HTML

  content_level_content_type_html
    -- Yes --> content_level_send_to_html
    --> content_level_content_type_html_score(Add to score)
    --> content_level_end

  content_level_content_type_html
    -- No --> content_level_content_type_js

  %% Check for JS

  content_level_content_type_js
    -- Yes --> content_level_content_type_js_external{Is external file?}
    -- Yes --> content_level_content_type_js_external_true(Request external 
                                                            file content)
    --> content_level_send_to_js
    --> content_level_content_type_js_score(Add to score)
    --> content_level_end

  content_level_content_type_js
    -- No --> content_level_end

  content_level_content_type_js_external
    -- No --> content_level_send_to_js
end

content_level_end
  --> flow_end