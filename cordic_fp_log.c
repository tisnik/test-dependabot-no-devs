// --------------------------------------------------------
// Výpočet hodnot funkce log() pomocí iteračního
// algoritmu CORDIC.
// --------------------------------------------------------

#include <math.h>
#include <stdio.h>
#include <stdlib.h>

// maximální počet iterací při běhu algoritmu
#define MAXITER 10

// "zesílení" při rotacích (odpovídá ln(2))
#define K 0.69314718056f

// ln(1+2*(-i))
double tabp[MAXITER] =
{ 
    0.40546510810816f,
    0.22314355131421f,
    0.11778303565638f,
    0.06062462181643f,
    0.03077165866675f,
    0.01550418653597f,
    0.00778214044205f,
    0.00389864041566f,
    0.00195122013126f,
    0.00097608597306f,
};

// ln(1-2*(-i))
double tabm[MAXITER] =
{
    -0.69314718055995f,
    -0.28768207245178f,
    -0.13353139262452f,
    -0.06453852113757f,
    -0.03174869831458f,
    -0.01574835696814f,
    -0.00784317746103f,
    -0.00391389932114f,
    -0.00195503483580f,
    -0.00097703964783f,
};

// výpočet logaritmu algoritmem CORDIC
double log_cordic(double a)
{
    const double three_eigth = 0.375f;
    int sk, expo;
    double sum = tabm[0];
    double x = 2.0f * frexpf (a, &expo);
    double ex2 = 1.0f; // dvojková mocnina
    int k;

    for (k = 0; k < MAXITER; k++) {
        sk = 0;
        // zjistit směr rotace
        if ((x - 1.0f) <  (-three_eigth * ex2)) {
            sk = +1;
        }
        if ((x - 1.0f) >= (+three_eigth * ex2)) {
            sk = -1;
        }
        ex2 /= 2.0;
        if (sk == 1) { // přímá rotace
            x = x + x * ex2;
            sum = sum - tabp [k];
        } 
        if (sk == -1) { // opačná rotace
            x = x - x * ex2;
            sum = sum - tabm [k];
        }
    }
    return expo * K + sum; // přepočet logaritmu
}


int main (void) {
    double a = M_E - 2.0; // "pěkná" počáteční hodnota

    for (; a<=2.0*M_E; a+=0.1) {
        double log_value = log_cordic(a);
        double log_correct = log(a);
        double log_error = fabs(log_correct - log_value);
        // tisk výsledků
        printf("%5.3f\t%12.10f\t%12.10f\t%8.3f%%\n",
               a,
               log_value,
               log_error,
               (log_value != 0.0) ? 100.0 * log_error / fabs(log_value) : 0.0);
    }
    return 0;
}

// finito
