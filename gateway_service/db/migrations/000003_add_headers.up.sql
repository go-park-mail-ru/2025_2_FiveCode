-- Insert header files into the file table
INSERT INTO file (url, mime_type, size_bytes, width, height, created_at, updated_at)
VALUES
    ('http://minio:9000/notes-app/headers/birds.jpg', 'image/jpeg', 50000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('http://minio:9000/notes-app/headers/carpet.jpg', 'image/jpeg', 50000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('http://minio:9000/notes-app/headers/earth.jpg', 'image/jpeg', 50000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('http://minio:9000/notes-app/headers/flowers.jpg', 'image/jpeg', 50000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('http://minio:9000/notes-app/headers/solid_blue.png', 'image/png', 10000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('http://minio:9000/notes-app/headers/solid_red.png', 'image/png', 10000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('http://minio:9000/notes-app/headers/solid_yellow.png', 'image/png', 10000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('http://minio:9000/notes-app/headers/stars.jpg', 'image/jpeg', 50000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('http://minio:9000/notes-app/headers/willow.jpg', 'image/jpeg', 50000, 1200, 600, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
