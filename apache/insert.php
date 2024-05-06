<?php
// Проверяем, является ли метод запроса POST
if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    // Если это не POST-запрос, мы отправляем ответ об ошибке.
    http_response_code(405); // метод не разрешен
    echo json_encode(['error' => 'This endpoint only accepts POST requests.']);
    exit;
}

// Читаем данные JSON из тела запроса
$jsonData = file_get_contents('php://input');

// Декодируем данные JSON в ассоциативный массив
$data = json_decode($jsonData, true);

// Проверяем, прошло ли декодирование JSON успешно
if ($data === null) {
    // мы отправляем ответ об ошибке, если JSON недействителен
    http_response_code(400);
    echo json_encode(['error' => 'Invalid JSON data']);
    exit;
}

$PathName = $data['PathName'] ?? null;
$Size = $data['Size'] ?? null;
$Elapsedtime = $data['ElapsedTime'] ?? null;

// Здесь мы  обрабатываем данные, 
error_log("Received data: PathName= " . $PathName . ", Size = " . $Size, ", Size = " . $ElapsedTime);

/*
$mysqli = new mysqli("localhost", "username", "password", "database");
if ($mysqli->connect_errno) {
    http_response_code(500);
    echo json_encode(['error' => 'Failed to connect to database: ' . $mysqli->connect_error]);
    exit;
}

$stmt = $mysqli->prepare("INSERT INTO STITISTICS (C_PATHNAME, C_SIZE, C_ELAPSEDTIME) VALUES (?, ?)");
$stmt->bind_param("sss", $PathName, $Size, $ElapsedTime);
$stmt->execute();
$stmt->close();
$mysqli->close();
*/

// If everything is successful, send a response
http_response_code(200);
echo json_encode(['message' => 'Data processed successfully']);
print_r($data);
exit;
