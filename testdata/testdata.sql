-- Insert Customer test data
INSERT INTO customers (id, name, email, created_at, updated_at)
VALUES 
    ('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'John Smith', 'john.smith@example.com', '2025-01-15 09:30:00'::timestamp, '2025-01-15 09:30:00'::timestamp),
    ('b2c3d4e5-f6a7-8901-bcde-f12345678901', 'Jane Doe', 'jane.doe@example.com', '2025-01-20 14:45:00'::timestamp, '2025-01-20 14:45:00'::timestamp),
    ('c3d4e5f6-a7b8-9012-cdef-123456789012', 'Michael Johnson', 'michael.johnson@example.com', '2025-02-05 11:15:00'::timestamp, '2025-02-05 11:15:00'::timestamp),
    ('d4e5f6a7-b8c9-0123-defa-234567890123', 'Sarah Williams', 'sarah.williams@example.com', '2025-02-10 16:20:00'::timestamp, '2025-02-10 16:20:00'::timestamp),
    ('e5f6a7b8-c9d0-1234-efab-345678901234', 'Robert Brown', 'robert.brown@example.com', '2025-03-01 10:00:00'::timestamp, '2025-03-01 10:00:00'::timestamp);

-- Insert Service Provider test data
INSERT INTO service_providers (id, name, email, created_at, updated_at)
VALUES 
    ('f6a7b8c9-d0e1-2345-fabc-456789012345', 'HomeClean Services', 'info@homeclean.example.com', '2025-01-10 08:00:00'::timestamp, '2025-01-10 08:00:00'::timestamp),
    ('a7b8c9d0-e1f2-3456-abcd-567890123456', 'Green Lawn Care', 'support@greenlawn.example.com', '2025-01-12 09:15:00'::timestamp, '2025-01-12 09:15:00'::timestamp),
    ('b8c9d0e1-f2a3-4567-bcde-678901234567', 'Electric Experts', 'service@electricexperts.example.com', '2025-01-25 10:30:00'::timestamp, '2025-01-25 10:30:00'::timestamp),
    ('c9d0e1f2-a3b4-5678-cdef-789012345678', 'Plumbing Professionals', 'help@plumbingpros.example.com', '2025-02-01 11:45:00'::timestamp, '2025-02-01 11:45:00'::timestamp),
    ('d0e1f2a3-b4c5-6789-defa-890123456789', 'Home Renovation Team', 'projects@homereno.example.com', '2025-02-15 13:00:00'::timestamp, '2025-02-15 13:00:00'::timestamp);

-- Insert Rating test data
INSERT INTO ratings (id, customer_id, service_provider_id, rating_value, comment, created_at, updated_at)
VALUES 
    ('e1f2a3b4-c5d6-7890-abcd-901234567890', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'f6a7b8c9-d0e1-2345-fabc-456789012345', 5, 'Excellent cleaning service, very thorough!', '2025-03-10 14:20:00'::timestamp, '2025-03-10 14:20:00'::timestamp),
    ('f2a3b4c5-d6e7-8901-bcde-012345678901', 'b2c3d4e5-f6a7-8901-bcde-f12345678901', 'a7b8c9d0-e1f2-3456-abcd-567890123456', 4, 'Good lawn service, but left some clippings.', '2025-03-15 15:30:00'::timestamp, '2025-03-15 15:30:00'::timestamp),
    ('a3b4c5d6-e7f8-9012-cdef-123456789012', 'c3d4e5f6-a7b8-9012-cdef-123456789012', 'b8c9d0e1-f2a3-4567-bcde-678901234567', 5, 'Fixed my electrical issues quickly and professionally.', '2025-03-20 16:45:00'::timestamp, '2025-03-20 16:45:00'::timestamp),
    ('b4c5d6e7-f8a9-0123-defa-234567890123', 'd4e5f6a7-b8c9-0123-defa-234567890123', 'c9d0e1f2-a3b4-5678-cdef-789012345678', 3, 'Plumbing work was okay, but they were late.', '2025-03-25 17:55:00'::timestamp, '2025-03-25 17:55:00'::timestamp),
    ('c5d6e7f8-a9b0-1234-efab-345678901234', 'e5f6a7b8-c9d0-1234-efab-345678901234', 'd0e1f2a3-b4c5-6789-defa-890123456789', 5, 'Amazing renovation work! Transformed our space.', '2025-03-30 18:10:00'::timestamp, '2025-03-30 18:10:00'::timestamp),
    ('d6e7f8a9-b0c1-2345-cdef-456789012345', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'c9d0e1f2-a3b4-5678-cdef-789012345678', 4, 'Good plumbing service overall.', '2025-04-05 09:25:00'::timestamp, '2025-04-05 09:25:00'::timestamp),
    ('e7f8a9b0-c1d2-3456-bcde-567890123456', 'b2c3d4e5-f6a7-8901-bcde-f12345678901', 'b8c9d0e1-f2a3-4567-bcde-678901234567', 2, 'Electrical work needed to be redone.', '2025-04-10 10:40:00'::timestamp, '2025-04-10 10:40:00'::timestamp);
