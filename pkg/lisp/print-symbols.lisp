;;
;; prints a list of symbols defined in the top-level namespace.
;;
(print
  (string-join
    (sort
      (map (lambda (x) (str x))
           (environment-bound-names (the-environment)))
      (lambda (x y) (string<? x y)))
    "\n"))