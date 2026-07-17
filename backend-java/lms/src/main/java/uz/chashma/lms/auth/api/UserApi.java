package uz.chashma.lms.auth.api;

import java.time.Instant;
import java.util.List;

/**
 * auth modulining public facade'i — boshqa modullar user ma'lumotini FAQAT
 * shu interfeys orqali oladi (Go'dagi GET /internal/users o'rni).
 * Microservice'ga ajratilganda bu interfeys REST/gRPC client bilan almashadi.
 */
public interface UserApi {

    List<UserSummary> findByIds(List<Long> ids);

    record UserSummary(long id, Instant createdAt, String name, String email, String role) {
    }
}
