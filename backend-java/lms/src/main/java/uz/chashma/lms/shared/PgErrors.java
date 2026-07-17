package uz.chashma.lms.shared;

import org.postgresql.util.PSQLException;
import org.postgresql.util.ServerErrorMessage;
import org.springframework.dao.DataAccessException;

/**
 * Go'dagi pq.Error.Constraint tekshiruvining ekvivalenti: Postgres
 * constraint nomini ajratib oladi (duplicate email/slug, FK va h.k.).
 */
public final class PgErrors {

    private PgErrors() {
    }

    public static String constraint(DataAccessException ex) {
        Throwable cause = ex;
        while (cause != null) {
            if (cause instanceof PSQLException psql) {
                ServerErrorMessage msg = psql.getServerErrorMessage();
                return msg == null ? null : msg.getConstraint();
            }
            cause = cause.getCause();
        }
        return null;
    }
}
