// NotAGame.cpp : This file contains the 'main' function. Program execution begins and ends there.
//

#include <iostream>
#include <chrono>
#include <thread>
#include <string>

int main(const int argc, const char *argv[]) {
    std::cout << "Hello from C++" << std::endl;

    const int total = argc - 1;
    for (int i = 1; i < argc; i++) {
        const int secondsToSleep = std::stoi(argv[i]);

        // If seconds to sleep is negative, print to the stderr stream. Else, print to the stdout stream.
        std::basic_ostream<char> *channel = secondsToSleep < 0 ? &std::cerr : &std::cout;

        // Before sleep message
        *channel << i << " of " << total << ": will sleep for" << secondsToSleep << std::endl;

        // Sleep
        std::this_thread::sleep_for(std::chrono::seconds(std::abs(secondsToSleep)));

        // After sleep message
        *channel << i << " of " << total << ": did sleep for" << secondsToSleep << std::endl;
    }

    std::cout << "Good bye from C++" << std::endl;
}

// Run program: Ctrl + F5 or Debug > Start Without Debugging menu
// Debug program: F5 or Debug > Start Debugging menu

// Tips for Getting Started:
//   1. Use the Solution Explorer window to add/manage files
//   2. Use the Team Explorer window to connect to source control
//   3. Use the Output window to see build output and other messages
//   4. Use the Error List window to view errors
//   5. Go to Project > Add New Item to create new code files, or Project > Add Existing Item to add existing code files to the project
//   6. In the future, to open this project again, go to File > Open > Project and select the .sln file
