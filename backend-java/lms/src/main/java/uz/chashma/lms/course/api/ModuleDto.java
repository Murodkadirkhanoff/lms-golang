package uz.chashma.lms.course.api;

import com.fasterxml.jackson.annotation.JsonIgnore;

import java.util.ArrayList;
import java.util.List;

public class ModuleDto {
    public long id;
    public String title;
    @JsonIgnore
    public int position;
    public List<LessonDto> lessons = new ArrayList<>();
}
