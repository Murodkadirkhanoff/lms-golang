package uz.chashma.lms.course;

import com.fasterxml.jackson.annotation.JsonIgnore;

import java.util.ArrayList;
import java.util.List;

/**
 * Frontend Quiz tipi bilan bir xil JSON. Baholash clientda bo'lgani uchun
 * correctIndex javobga kiradi (Go'dagi kabi).
 */
class QuizDto {
    public long id;
    @JsonIgnore
    public long courseId;
    public String title;
    public int passingScore;
    public int timeLimitMinutes;
    public List<QuestionDto> questions = new ArrayList<>();
    @JsonIgnore
    public int version;

    static class QuestionDto {
        public long id;
        public String question;
        public List<String> options = new ArrayList<>();
        public int correctIndex;
        @JsonIgnore
        public int position;
    }
}
