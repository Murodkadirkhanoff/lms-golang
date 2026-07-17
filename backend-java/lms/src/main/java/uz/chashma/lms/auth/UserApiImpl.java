package uz.chashma.lms.auth;

import org.springframework.stereotype.Service;
import uz.chashma.lms.auth.api.UserApi;

import java.util.List;

@Service
class UserApiImpl implements UserApi {

    private final UserRepository users;

    UserApiImpl(UserRepository users) {
        this.users = users;
    }

    @Override
    public List<UserSummary> findByIds(List<Long> ids) {
        if (ids == null || ids.isEmpty()) {
            return List.of();
        }
        return users.findByIds(ids).stream()
                .map(u -> new UserSummary(u.id, u.createdAt, u.name, u.email, u.role))
                .toList();
    }
}
