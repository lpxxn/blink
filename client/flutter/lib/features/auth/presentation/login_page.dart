import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../application/auth_state_provider.dart';
import '../application/login_submit_provider.dart';
import '../data/auth_repository.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();
  final TextEditingController _email = TextEditingController();
  final TextEditingController _password = TextEditingController();

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    ref.listen(loginSubmitProvider, (AsyncValue<void>? previous, AsyncValue<void> next) {
      if (!next.hasError) {
        return;
      }
      final Object e = next.error!;
      final String msg;
      if (e is AuthInvalidCredentialsException) {
        msg = 'Invalid email or password';
      } else if (e is AuthRepositoryException) {
        msg = e.message;
      } else {
        msg = 'Login failed';
      }
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    });

    final AsyncValue<void> submitAsync = ref.watch(loginSubmitProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Login')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              TextFormField(
                controller: _email,
                decoration: const InputDecoration(labelText: 'Email'),
                keyboardType: TextInputType.emailAddress,
                autofillHints: const [AutofillHints.email],
                validator: (String? v) =>
                    (v == null || v.isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _password,
                decoration: const InputDecoration(labelText: 'Password'),
                obscureText: true,
                autofillHints: const [AutofillHints.password],
                validator: (String? v) =>
                    (v == null || v.isEmpty) ? 'Required' : null,
              ),
              const SizedBox(height: 20),
              FilledButton(
                onPressed: submitAsync.isLoading
                    ? null
                    : () async {
                        if (!(_formKey.currentState?.validate() ?? false)) {
                          return;
                        }
                        await ref.read(loginSubmitProvider.notifier).submit(
                              _email.text.trim(),
                              _password.text,
                            );
                        if (!context.mounted) {
                          return;
                        }
                        if (ref.read(authStateProvider) ==
                            AuthStatus.authenticated) {
                          context.go('/');
                        }
                      },
                child: const Text('Sign in'),
              ),
              TextButton(
                onPressed: () => context.go('/'),
                child: const Text('Back'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
