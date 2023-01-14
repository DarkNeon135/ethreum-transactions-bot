That bot was created to receive ethereum transactions with transfer mode from blocks range.
To use it specify mode in cli arguments:
0 - bot will use RPC API endpoints from the providers list array and swap it time to time;
leave blank to use 1 RPC API endpoint or local blockchain files.
Always you should determine start block and end block inside the code.
Transactions writes to directory /assets, if program will be interrupted you can continue it at any time, because bot has save and restore function.
Good luck!
