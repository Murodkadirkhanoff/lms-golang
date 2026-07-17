package uz.chashma.lms.shared;

public class EditConflictException extends RuntimeException {
    public EditConflictException() {
        super("unable to update the record due to an edit conflict, please try again");
    }
}
