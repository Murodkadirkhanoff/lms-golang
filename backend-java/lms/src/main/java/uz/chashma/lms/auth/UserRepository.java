package uz.chashma.lms.auth;

import org.springframework.dao.DataAccessException;
import org.springframework.jdbc.core.RowMapper;
import org.springframework.jdbc.core.simple.JdbcClient;
import org.springframework.stereotype.Repository;
import uz.chashma.lms.shared.EditConflictException;
import uz.chashma.lms.shared.NotFoundException;
import uz.chashma.lms.shared.PgErrors;

import java.sql.ResultSet;
import java.sql.SQLException;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Optional;

/** Go auth-service/internal/data/users.go porti (SQL 1:1). */
@Repository
class UserRepository {

    static class DuplicateEmailException extends RuntimeException {
    }

    private final JdbcClient jdbc;

    UserRepository(JdbcClient jdbc) {
        this.jdbc = jdbc;
    }

    private static final RowMapper<User> FULL_MAPPER = (rs, rowNum) -> {
        User user = base(rs);
        user.passwordHash = rs.getBytes("password_hash");
        user.version = rs.getInt("version");
        return user;
    };

    private static User base(ResultSet rs) throws SQLException {
        User user = new User();
        user.id = rs.getLong("id");
        user.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
        user.name = rs.getString("name");
        user.email = rs.getString("email");
        user.role = rs.getString("role");
        return user;
    }

    void insert(User user) {
        try {
            jdbc.sql("""
                    INSERT INTO auth.users (name, email, password_hash, role)
                    VALUES (:name, :email, :hash, :role)
                    RETURNING id, created_at, version
                    """)
                    .param("name", user.name)
                    .param("email", user.email)
                    .param("hash", user.passwordHash)
                    .param("role", user.role)
                    .query((org.springframework.jdbc.core.RowCallbackHandler) rs -> {
                        user.id = rs.getLong("id");
                        user.createdAt = rs.getObject("created_at", OffsetDateTime.class).toInstant();
                        user.version = rs.getInt("version");
                    });
        } catch (DataAccessException ex) {
            if ("users_email_key".equals(PgErrors.constraint(ex))) {
                throw new DuplicateEmailException();
            }
            throw ex;
        }
    }

    Optional<User> findByEmail(String email) {
        return jdbc.sql("""
                SELECT id, created_at, name, email, password_hash, role, version
                FROM auth.users
                WHERE email = :email AND deleted_at IS NULL
                """)
                .param("email", email)
                .query(FULL_MAPPER)
                .optional();
    }

    Optional<User> findById(long id) {
        if (id < 1) {
            return Optional.empty();
        }
        return jdbc.sql("""
                SELECT id, created_at, name, email, password_hash, role, version
                FROM auth.users
                WHERE id = :id AND deleted_at IS NULL
                """)
                .param("id", id)
                .query(FULL_MAPPER)
                .optional();
    }

    List<User> findByIds(List<Long> ids) {
        return jdbc.sql("""
                SELECT id, created_at, name, email, role
                FROM auth.users
                WHERE id IN (:ids) AND deleted_at IS NULL
                """)
                .param("ids", ids)
                .query((rs, rowNum) -> base(rs))
                .list();
    }

    void updatePassword(User user) {
        int updated = jdbc.sql("""
                UPDATE auth.users
                SET password_hash = :hash, version = version + 1
                WHERE id = :id AND version = :version
                """)
                .param("hash", user.passwordHash)
                .param("id", user.id)
                .param("version", user.version)
                .update();
        if (updated == 0) {
            throw new EditConflictException();
        }
        user.version++;
    }

    /** Profil: ismni yangilaydi (optimistic locking bilan). */
    void updateName(User user) {
        int updated = jdbc.sql("""
                UPDATE auth.users
                SET name = :name, version = version + 1
                WHERE id = :id AND version = :version
                """)
                .param("name", user.name)
                .param("id", user.id)
                .param("version", user.version)
                .update();
        if (updated == 0) {
            throw new EditConflictException();
        }
        user.version++;
    }

    /** Admin panel: user rolini almashtiradi. */
    void updateRole(long id, String role) {
        int updated = jdbc.sql("""
                UPDATE auth.users
                SET role = :role, version = version + 1
                WHERE id = :id AND deleted_at IS NULL
                """)
                .param("role", role)
                .param("id", id)
                .update();
        if (updated == 0) {
            throw new NotFoundException();
        }
    }

    int count() {
        return jdbc.sql("SELECT count(*) FROM auth.users WHERE deleted_at IS NULL")
                .query(Integer.class)
                .single();
    }

    record UserPage(List<User> users, int total) {
    }

    UserPage list(int page, int pageSize) {
        var totalHolder = new int[1];
        List<User> users = jdbc.sql("""
                SELECT count(*) OVER() AS total, id, created_at, name, email, role
                FROM auth.users
                WHERE deleted_at IS NULL
                ORDER BY id
                LIMIT :limit OFFSET :offset
                """)
                .param("limit", pageSize)
                .param("offset", (page - 1) * pageSize)
                .query((rs, rowNum) -> {
                    totalHolder[0] = rs.getInt("total");
                    return base(rs);
                })
                .list();
        return new UserPage(users, totalHolder[0]);
    }
}
