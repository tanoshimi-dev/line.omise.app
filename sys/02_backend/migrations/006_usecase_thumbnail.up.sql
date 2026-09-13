ALTER TABLE usecases
    ADD COLUMN thumbnail_url TEXT,
    ADD COLUMN related_demo_app TEXT
        CHECK (related_demo_app IN ('membership', 'salon-reservation', 'sweets-shop'));
