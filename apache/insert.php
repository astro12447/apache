<?php
// Check if the request method is POST
if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    // If not a POST request, send an error response
    http_response_code(405); // Method Not Allowed
    echo json_encode(['error' => 'This endpoint only accepts POST requests.']);
    exit;
}

// Read the JSON data from the request body
$jsonData = file_get_contents('php://input');

// Decode the JSON data into an associative array
$data = json_decode($jsonData, true);

// Check if the JSON decode was successful
if ($data === null) {
    // Send an error response if the JSON is not valid
    http_response_code(400);
    echo json_encode(['error' => 'Invalid JSON data']);
    exit;
}

// Assuming the JSON data contains fields 'field1' and 'field2'
$field1 = $data['field1'] ?? null;
$field2 = $data['field2'] ?? null;

// Here you would typically process the data, 
error_log("Received data: field1 = " . $field1 . ", field2 = " . $field2);

/*
$mysqli = new mysqli("localhost", "username", "password", "database");
if ($mysqli->connect_errno) {
    http_response_code(500);
    echo json_encode(['error' => 'Failed to connect to database: ' . $mysqli->connect_error]);
    exit;
}

$stmt = $mysqli->prepare("INSERT INTO STATISTICS (field1, field2) VALUES (?, ?)");
$stmt->bind_param("ss", $field1, $field2);
$stmt->execute();
$stmt->close();
$mysqli->close();
*/

// If everything is successful, send a response
http_response_code(200);
echo json_encode(['message' => 'Data processed successfully']);
print_r($data);
exit;
