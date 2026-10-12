package mindconnect

import "testing"

func TestValue(t *testing.T) {
	if n, err := Value("-2").Int(); err != nil || n != -2 {
		t.Errorf("Int = %d, err = %v, ожидалось -2", n, err)
	}
	if _, err := Value("").Int(); err == nil {
		t.Error("Int пустого значения: ожидалась ошибка")
	}
	if b, err := Value("True").Bool(); err != nil || !b {
		t.Errorf("Bool = %v, err = %v, ожидалось true", b, err)
	}
	if _, err := Value("true").Bool(); err == nil {
		t.Error("Bool учитывает регистр: ожидалась ошибка")
	}
	for _, code := range []Value{"cmd_err", "prt_err", "fnc_err0", "fnc_err1", "fnc_err2", "fnc_err3", "fnc_err9"} {
		if !code.IsError() {
			t.Errorf("%s: IsError = false", code)
		}
	}
	if Value("cmd_done").IsError() || Value("80").IsError() {
		t.Error("IsError = true для обычного значения")
	}
	for _, v := range []Value{"fnc_err", "fnc_err12"} {
		if v.IsError() {
			t.Errorf("%q: IsError = true, ожидалось, что неполный код останется обычным значением", v)
		}
	}
}

func TestResponseReaderSentinelIsError(t *testing.T) {
	r := NewResponseReader(7, 3)
	done, err := r.Feed([]byte("7\x1f80\x1frocky_plateau\x1ffnc_err1"))
	if !done || err != nil {
		t.Fatalf("done=%v err=%v, ожидалось, что ответ завершён", done, err)
	}
	id, values, err := r.Response()
	if err != nil || id != 7 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	if !values[2].IsError() {
		t.Errorf("values = %q, ожидалось, что ответ сторожевого запроса отмечен как ошибка", values)
	}
}

func TestResponseReaderUnknownErrorSentinel(t *testing.T) {
	r := NewResponseReader(7, 2)
	if done, err := r.Feed([]byte("7\x1f80\x1ffnc_err9")); !done || err != nil {
		t.Fatalf("done=%v err=%v, ожидалось, что неизвестный fnc_err завершит ответ", done, err)
	}
}
