package utils

func SafeStringDeref(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

func SafeInt32Deref(i *int32) int32 {
    if i == nil {
        return 0
    }
    return *i
}