# описание работы алгоритма

1. убираются все объявления которые не попадают под условия таргетинга
2. убираются все объявления которые пользователь уже видел (чтобы просмотр был уникальным)
3. для каждого объявления вычисляется score по следующей формуле:

```math
\text{score} = \left(
0.5 \times \frac{\text{cost\_per\_impression} - \operatorname{avg}(\text{cost\_per\_impression})}{\operatorname{stddev}(\text{cost\_per\_impression})} +
0.3 \times \frac{\text{cost\_per\_click} - \operatorname{avg}(\text{cost\_per\_click})}{\operatorname{stddev}(\text{cost\_per\_click})} +
0.2 \times \frac{\text{ml\_score} - \operatorname{avg}(\text{ml\_score})}{\operatorname{stddev}(\text{ml\_score})}
\right) \times \left( 1 - \left(\frac{\text{impressions\_count}}{\text{impressions\_limit}}\right)^3 \right)

```

4. выбирается объявление с наибольшим score
