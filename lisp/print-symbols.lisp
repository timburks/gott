;;
;; prints a list of symbols defined in the top-level namespace.
;;
(write-line
  (string-join
    (sort
      (map (lambda (x) (str x))
           (environment-bound-names (the-environment)))
      (lambda (x y) (string<? x y)))
    "\n"))
