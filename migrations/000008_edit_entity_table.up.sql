ALTER TABLE entries ADD image_link TEXT;
ALTER TABLE entries ADD entry_link TEXT;
ALTER TABLE entries ADD entry_type TEXT CHECK(entry_type IN ('youtube', 'animethemes'));