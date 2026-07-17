package uz.chashma.lms.shared;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

@RestController
public class HealthController {

    private static final String VERSION = "1.0.0";

    private final LmsProperties props;

    public HealthController(LmsProperties props) {
        this.props = props;
    }

    @GetMapping("/v1/healthcheck")
    public Map<String, Object> healthcheck() {
        return Map.of(
                "status", "available",
                "system_info", Map.of(
                        "service", "lms",
                        "environment", props.env(),
                        "version", VERSION));
    }
}
