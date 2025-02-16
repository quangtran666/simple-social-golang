ALTER TABLE
    user_invitations
ADD
    COLUMN expiry timestamp(0) with time zone not null;