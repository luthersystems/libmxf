import { Mxf } from '../index';

test('decode', async () => {
  const mxf = await new Mxf().initialized;

  function getKey(dsid: string): string {
    const keys: { [dsid: string]: string } = {
      '89871b002200cbd38dec6cc43988e687f0a6799e2734635fee7cc891eca1170d-1': '22c20ac7e93194e19661cb55fc3633bbfd1a44c82d852fc8ee5e160aaba7c9a0',
    };
    return keys[dsid];
  }

  const encMsg = JSON.stringify({
    message: null,
    mxf: 'v1',
    transforms: [
      {
        body: {
          dsid: '89871b002200cbd38dec6cc43988e687f0a6799e2734635fee7cc891eca1170d-1',
          encrypted_base64:
            'Y4vQZSIr88swK2GTTZR65v7jh/q6Z9VtITJ1MAcShAg5oqvL5dHBz8Iy4u+I7D8rHJHTiGY9MyQnJGQC6OE6WPkA0UQJAuEIyWnrnBCagMBnNG0qFfcmQ2/GdY7sI3oWTATA5La2KkYd662y2n1Jig9R7QKeOo8d0ANnKqyn2qsfBOXV6qnvAWvbOTOsxnMYrZ/VpQC6F+w=',
        },
        context_path: '.',
        header: { compressor: 'zlib', encryptor: 'AES-256', private_paths: ['.'], profile_paths: ['.client_id'] },
      },
    ],
  });

  expect(await mxf.decode(getKey, encMsg)).toBe('{"client_id":"dba7bd5f-7a8f-4797-9851-202885837845","client_originator":"cbfc1ad7-84ec-4a4b-8302-7b869c00bd8f","field_mask":["client_originator"]}');
});
