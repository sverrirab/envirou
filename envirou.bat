@FOR /F "tokens=* USEBACKQ" %%g IN (`py %0\..\envirou.py %*`) do @%%g