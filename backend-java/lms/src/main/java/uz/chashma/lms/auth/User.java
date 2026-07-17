package uz.chashma.lms.auth;

import java.time.Instant;

/** auth.users qatori. Modul ichida ishlatiladi — tashqariga UserApi chiqadi. */
class User {
    long id;
    Instant createdAt;
    String name;
    String email;
    byte[] passwordHash;
    String role;
    int version;
}
