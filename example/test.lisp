(+ 1 2)
(define foo (lambda (a b) (+ a b)))
(foo 4 5)
(define factorial
  (lambda (n)
    (if (= n 0) 1
        (* n (factorial (- n 1))))))
(factorial 100)
(car (list 1 2 3))
(cdr (list 1 2 3))
