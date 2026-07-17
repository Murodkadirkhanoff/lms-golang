package uz.chashma.lms.shared.security;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatus;
import org.springframework.security.config.Customizer;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;
import uz.chashma.lms.shared.JsonErrorWriter;
import uz.chashma.lms.shared.LmsProperties;

import java.util.List;

@Configuration
@EnableWebSecurity
public class SecurityConfig {

    @Bean
    public JwtService jwtService(LmsProperties props) {
        String secret = props.jwt().secret();
        if (secret == null || secret.isBlank()) {
            throw new IllegalStateException("JWT_SECRET must be set (at least 32 bytes for HS256)");
        }
        return new JwtService(secret, props.jwt().ttl());
    }

    // Go bilan bir xil: bcrypt cost 12 ($2a$) — hashlar o'zaro mos.
    @Bean
    public PasswordEncoder passwordEncoder() {
        return new BCryptPasswordEncoder(12);
    }

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http, JwtService jwtService,
                                           LmsProperties props) throws Exception {
        LmsProperties.RateLimit rl = props.rateLimit();
        http
                .csrf(AbstractHttpConfigurer::disable)
                .cors(Customizer.withDefaults())
                .sessionManagement(s -> s.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                .exceptionHandling(e -> e
                        // Go RequireAuthenticated / RequireRole javoblari
                        .authenticationEntryPoint((req, res, ex) ->
                                JsonErrorWriter.write(res, HttpStatus.UNAUTHORIZED.value(),
                                        "you must be authenticated to access this resource"))
                        .accessDeniedHandler((req, res, ex) ->
                                JsonErrorWriter.write(res, HttpStatus.FORBIDDEN.value(),
                                        "your user account doesn't have the necessary permissions to access this resource")))
                .addFilterBefore(new JwtAuthFilter(jwtService), UsernamePasswordAuthenticationFilter.class)
                // Brute-force himoyasi: auth endpointlariga IP bo'yicha limit
                .addFilterBefore(new RateLimitFilter(rl.enabled(), rl.rps(), rl.burst()),
                        JwtAuthFilter.class)
                .authorizeHttpRequests(a -> a
                        // Go route guruhlari bilan bir xil himoya
                        .requestMatchers("/v1/admin/**").hasRole(Roles.ADMIN)
                        .requestMatchers(HttpMethod.POST, "/v1/courses").authenticated()
                        .requestMatchers(HttpMethod.PATCH, "/v1/courses/*").authenticated()
                        .requestMatchers(HttpMethod.DELETE, "/v1/courses/*").authenticated()
                        .requestMatchers(HttpMethod.POST, "/v1/lessons/*/questions").authenticated()
                        .requestMatchers("/v1/courses/*/reviews", "/v1/courses/*/quiz",
                                "/v1/quizzes/*/attempts", "/v1/courses/*/enroll",
                                "/v1/enrollments/**", "/v1/me/**", "/v1/uploads").authenticated()
                        .anyRequest().permitAll());
        return http.build();
    }

    @Bean
    public CorsConfigurationSource corsConfigurationSource(LmsProperties props) {
        CorsConfiguration config = new CorsConfiguration();
        config.setAllowedOrigins(List.of(props.corsTrustedOrigins().trim().split("\\s+")));
        config.setAllowedMethods(List.of("GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"));
        config.setAllowedHeaders(List.of("Authorization", "Content-Type"));

        UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
        source.registerCorsConfiguration("/**", config);
        return source;
    }
}
